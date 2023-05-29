package queue

import (
	"github.com/AleksandrMac/testfsd/internal/entities"
	"github.com/AleksandrMac/testfsd/pkg/ws"
)

type ChanSliceByte struct {
	Ch    chan []byte
	Close func()
}

type Queuer interface {
	NewMessageListener(rooms []entities.Room, conn *ws.WS) ChanSliceByte
	PushMessage(msg *entities.Message, skpipConn *ws.WS) error
	PushNotification(notification *entities.Notification, skpipConn *ws.WS) error
	Close()
}
