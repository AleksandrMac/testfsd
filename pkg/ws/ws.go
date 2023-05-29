package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WS struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func NewWS(conn *websocket.Conn) *WS {
	return &WS{conn: conn}
}

func (x *WS) SendMessage(msg []byte) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.conn.WriteMessage(websocket.TextMessage, msg)
}
