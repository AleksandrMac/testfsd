package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/AleksandrMac/testfsd/internal/config"
	"github.com/AleksandrMac/testfsd/internal/entities"
	"github.com/AleksandrMac/testfsd/internal/log"
	"github.com/AleksandrMac/testfsd/pkg/ws"

	amqp "github.com/rabbitmq/amqp091-go"
)

type queueRabbit struct {
	ctx      context.Context
	rmqconn  *amqp.Connection
	rmpPubCh *amqp.Channel
	// rmpQueue  amqp.Queue
	topicName string
	listeners *sync.Map
	conns     map[chan []byte]*ws.WS
	closeFn   []func()
}

// NewOperationDatastoreExternal get operation external api
func NewQueueRabbitMQ(ctx context.Context, conf config.RabbitMQ) (Queuer, error) {
	var err error
	qName := "msg"

	q := queueRabbit{
		listeners: &sync.Map{},
		conns:     make(map[chan []byte]*ws.WS),
		topicName: qName,
	}
	// conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	q.rmqconn, err = amqp.Dial(conf.DSN)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %w", err)
	}
	q.closeFn = append(q.closeFn, func() {
		if err := q.rmqconn.Close(); err != nil {
			log.Default().Warn("failed to close rabbitMQ connect: " + err.Error())
		}
	})

	if q.rmpPubCh, err = q.rmqconn.Channel(); err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}
	q.closeFn = append(q.closeFn, func() {
		if err := q.rmpPubCh.Close(); err != nil {
			log.Default().Warn("failed to close rabbitMQ channel " + err.Error())
		}
	})

	q.rmpPubCh.ExchangeDeclare(
		q.topicName, // name
		"fanout",    // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)

	msgs, err := q.consumer()
	if err != nil {
		return nil, fmt.Errorf("failed init consumer: %w", err)
	}

	go func() {
		tick := time.NewTicker(5 * time.Second)
		defer log.Default().Info("close stub message listener")
		defer tick.Stop()
		for {
			select {
			case msg := <-msgs:
				log.Default().Info("Received a message: " + string(msg.Body))
				// ch <- d.Body
				// dotCount := bytes.Count(d.Body, []byte("."))
				// t := time.Duration(dotCount)
				// time.Sleep(t * time.Second)
				// log.Printf("Done")
				// msg.Ack(false)
				var resp entities.Response
				if err := json.Unmarshal(msg.Body, &resp); err != nil {
					log.Default().Warn("failed parse message: " + err.Error())
				}
				switch {
				case resp.Message != nil:
					value, ok := q.listeners.Load(resp.Message.RoomId)
					if !ok {
						continue
					}
					for ch := range value.(map[chan []byte]*ws.WS) {
						// вопрос как в раббит передать соединение которое нужно игнорировать при передаче сообщения
						// if conn == skipConn {
						// 	continue
						// }
						ch <- msg.Body
					}
				case resp.Notification != nil:
					sendedCh := map[chan []byte]struct{}{}
					q.listeners.Range(func(key, value any) bool {
						for ch := range value.(map[chan []byte]*ws.WS) {
							if _, ok := sendedCh[ch]; ok {
								continue
							}
							ch <- msg.Body
							sendedCh[ch] = struct{}{}
						}
						return true
					})
				}
			case <-tick.C:
				if config.Default.Server.Port != 5000 {
					continue
				}
				m := entities.Response{
					Message: &entities.Message{
						Text:   "stub message: " + time.Now().String(),
						RoomId: 1,
					},
				}
				msg, _ := json.Marshal(m)
				if err = q.rmpPubCh.PublishWithContext(ctx,
					q.topicName, // exchange
					"",          // routing key
					false,       // mandatory
					false,
					amqp.Publishing{
						ContentType: "application/json",
						Body:        msg,
					}); err != nil {
					log.Default().Warn("Failed to publish a message: " + err.Error())
				}
				log.Default().Info(string(msg))

			case <-ctx.Done():
				return
			}
		}
	}()

	return &q, nil
}

func (x *queueRabbit) Close() {
	for _, fn := range x.closeFn {
		fn()
	}
}

func (x *queueRabbit) consumer() (<-chan amqp.Delivery, error) {
	rabbitch, err := x.rmqconn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to open a channel: %w", err)
	}
	x.closeFn = append(x.closeFn, func() {
		if err := rabbitch.Close(); err != nil {
			log.Default().Warn("failed to close channel: " + err.Error())
		}
	})

	err = rabbitch.ExchangeDeclare(
		x.topicName, // name
		"fanout",    // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare an exchange: %w", err)
	}

	q, err := rabbitch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare a queue: %w", err)
	}

	if err = rabbitch.QueueBind(
		q.Name,      // queue name
		"",          // routing key
		x.topicName, // exchange
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("Failed to bind a queue: %w", err)
	}

	return rabbitch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
}

func (x *queueRabbit) NewMessageListener(listenRooms []entities.Room, conn *ws.WS) ChanSliceByte {
	ch := make(chan []byte)
	closeFn := []func(){}

	for i := range listenRooms {
		roomId := listenRooms[i].ID
		val, ok := x.listeners.LoadOrStore(roomId, map[chan []byte]*ws.WS{ch: conn})
		if ok {
			val.(map[chan []byte]*ws.WS)[ch] = conn
		}
		closeFn = append(closeFn, func() {
			m := val.(map[chan []byte]*ws.WS)
			delete(m, ch)
			if len(m) == 0 {
				x.listeners.Delete(roomId)
			}
		})
	}
	x.conns[ch] = conn
	closeFn = append(closeFn, func() {
		delete(x.conns, ch)
	})

	return ChanSliceByte{
		Ch: ch,
		Close: func() {
			close(ch)
			for i := range closeFn {
				closeFn[i]()
			}
		},
	}
}

func (x *queueRabbit) PushMessage(m *entities.Message, skipConn *ws.WS) error {
	msg, err := json.Marshal(entities.Response{Message: m})
	if err != nil {
		return fmt.Errorf("failed parse message before push: %w", err)
	}
	return x.rmpPubCh.Publish(
		x.topicName, // exchange
		"",          // routing key
		false,       // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         msg,
		})
}

func (x *queueRabbit) PushNotification(n *entities.Notification, skipConn *ws.WS) error {
	notification, err := json.Marshal(entities.Response{Notification: n})
	if err != nil {
		return fmt.Errorf("failed parse message before push: %w", err)
	}
	return x.rmpPubCh.Publish(
		x.topicName, // exchange
		"",          // routing key
		false,       // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         notification,
		})
}
