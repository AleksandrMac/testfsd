package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/AleksandrMac/testfsd/internal/datastores"
	"github.com/AleksandrMac/testfsd/internal/entities"
	"github.com/AleksandrMac/testfsd/internal/log"
	"github.com/AleksandrMac/testfsd/internal/queue"
	"github.com/AleksandrMac/testfsd/pkg/ws"
)

// OperationService api controller of produces
type MessageService interface {
	MessageSending(ctx context.Context, listenRooms []entities.Room, conn *ws.WS)
	GetRooms(ctx context.Context, userId int64) ([]entities.Room, error)
	GetMessages(ctx context.Context, roomId int64, afterMessageId uint64) ([]entities.Message, error)
	WriteMessage(ctx context.Context, msg *entities.Message, skipConn *ws.WS) (int64, error)
	Notification(ctx context.Context, notification *entities.Notification, skipConn *ws.WS) error
}

type messageService struct {
	queue   queue.Queuer
	storage datastores.CommonDatastore
	mu      sync.Mutex
}

// NewOperationService get operation service instance
func NewMessageService(q queue.Queuer, ds datastores.CommonDatastore) MessageService {
	return &messageService{
		queue:   q,
		storage: ds,
	}
}

func (x *messageService) MessageSending(ctx context.Context, listenRooms []entities.Room, conn *ws.WS) {
	ch := x.queue.NewMessageListener(listenRooms, conn)
	defer ch.Close()
	var err error
	for {
		select {
		case msg := <-ch.Ch:
			log.Default().Info(string(msg))
			if err = conn.SendMessage(msg); err != nil {
				log.Default().Warn("failed send queuer message: " + err.Error())
			}
		case <-ctx.Done():
			return
		}
	}
}

func (x *messageService) GetRooms(ctx context.Context, userId int64) ([]entities.Room, error) {
	return x.storage.GetRooms(ctx, userId)
}

func (x *messageService) GetMessages(ctx context.Context, roomId int64, afterMessageId uint64) ([]entities.Message, error) {
	return x.storage.GetMessages(ctx, roomId, afterMessageId)
}

func (x *messageService) WriteMessage(ctx context.Context, msg *entities.Message, skipConn *ws.WS) (int64, error) {
	id, err := x.storage.WriteMessage(ctx, msg)
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}
	if err = x.queue.PushMessage(msg, skipConn); err != nil {
		log.Default().Warn("can't send message to queue: " + err.Error())
	}
	return id, nil
}

func (x *messageService) Notification(ctx context.Context, notification *entities.Notification, skipConn *ws.WS) error {
	return x.queue.PushNotification(notification, skipConn)
}
