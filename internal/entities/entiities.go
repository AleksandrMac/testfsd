package entities

import (
	"time"
)

type Message struct {
	Model
	RoomId int64  `json:"room_id"`
	Text   string `json:"text"`
	Author User   `json:"author"`
} //@name Message

type Subscribe struct {
	Channel Channel `json:"channel"`
} //@name Subscribe

type Method struct {
	Authorization   *Authorization `json:"authorization"`
	ReadMessageList *struct {
		Message
		AfterMessageId uint64
	} `json:"read_message"`
	WriteMessage *Message `json:"write_message"`
	Rooms        *string  `json:"rooms"`
} //@name Method

type Authorization struct {
	UserId int64 `json:"user_id"`
} //@name Authorization

type User struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
} //@name Author

// // Model base model definition, including fields `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`, which could be embedded in your models
// //
type Model struct {
	ID        int64      `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"-"`
}

type Notification struct {
	Event    string                 `json:"event"`
	Metadata map[string]interface{} `json:"metadata"`
}

type Room struct {
	Model
	Name                         string `json:"name"`
	LastMessageDarutionInMinutes int64  `json:"-"`
}

type Response struct {
	Message      *Message      `json:"message,omitempty"`
	Notification *Notification `json:"notification,omitempty"`
}
