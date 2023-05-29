package datastores

import (
	"context"

	"github.com/AleksandrMac/testfsd/internal/entities"
)

// CommonStorage datastorage
type CommonDatastore interface {
	// Operation creates an operation and returns the operation id or an error
	GetMessages(ctx context.Context, roomId int64, afterMessageId uint64) ([]entities.Message, error)
	GetRooms(ctx context.Context, userId int64) ([]entities.Room, error)
	WriteMessage(context.Context, *entities.Message) (int64, error)
}
