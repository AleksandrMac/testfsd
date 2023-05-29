package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/AleksandrMac/testfsd/internal/entities"
	"github.com/AleksandrMac/testfsd/internal/log"
	"github.com/AleksandrMac/testfsd/pkg/ws"
)

type queueStub struct {
	ctx       context.Context
	listeners *sync.Map
	conns     map[chan []byte]*ws.WS
}

// NewOperationDatastoreExternal get operation external api
func NewQueueStub(ctx context.Context) (Queuer, error) {
	m := sync.Map{}

	go func() {
		tick := time.NewTicker(5 * time.Second)
		defer log.Default().Info("close stub message listener")
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				msg := []byte("stub message: " + time.Now().String())
				m.Range(func(key, value any) bool {
					room_id := key.(int64)
					for ch := range value.(map[chan []byte]*ws.WS) {
						ch <- append([]byte{}, append(msg, []byte(fmt.Sprintf(" room_id: %d", room_id))...)...)
					}
					return true
				})
			case <-ctx.Done():
				return
			}
		}
	}()

	return &queueStub{
		ctx:       ctx,
		listeners: &m,
		conns:     map[chan []byte]*ws.WS{},
	}, nil
}

func (x *queueStub) Close() {}

func (x *queueStub) NewMessageListener(listenRooms []entities.Room, conn *ws.WS) ChanSliceByte {
	ch := make(chan []byte)
	closeFn := make([]func(), 0, len(listenRooms))

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

func (x *queueStub) PushMessage(msg *entities.Message, skipConn *ws.WS) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed parse message before push: %w", err)
	}

	value, ok := x.listeners.Load(msg.RoomId)
	if !ok {
		return nil
	}

	chans := value.(map[chan []byte]*ws.WS)
	for ch, conn := range chans {
		if conn == skipConn {
			continue
		}
		ch <- msgBytes
	}

	return err
}

func (x *queueStub) PushNotification(notification *entities.Notification, skipConn *ws.WS) error {
	msgBytes, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed parse message before push: %w", err)
	}

	for ch, conn := range x.conns {
		if conn == skipConn {
			continue
		}
		ch <- msgBytes
	}

	return err
}
