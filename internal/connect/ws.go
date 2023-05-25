package connect

import (
	"net/http"
	"time"

	_ "github.com/AleksandrMac/testfsd/internal/entities"
	"github.com/AleksandrMac/testfsd/internal/services"

	"github.com/gorilla/websocket"
)

type connectorWS struct {
	service services.MessageService
}

// NewConnector get Connect service instance
func NewConnectorWS() Connector {
	return &connectorWS{service: services.NewMessageService()}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Пропускаем любой запрос
	},
}

// Listen godoc
// @Title ConnectWS
// @Summary Подключение к websocket
// @Description Подключение к websocket
// @ID connect-ws
// @Accept  json
// @Produce  json
// @Param request body entities.Message true "operation info"
// @Success 204
// @Failure 400 {object} dto.ErrorReply "Bad Request"
// @Failure 500 {object} dto.ErrorReply "Unknown error"
// @Router /api/v1/operation [post]
func (x *connectorWS) Listen(rw http.ResponseWriter, r *http.Request) {
	connection, _ := upgrader.Upgrade(rw, r, nil)
	defer connection.Close() // Закрываем соединение
	for {
		connection.WriteMessage(websocket.TextMessage, []byte("hello world"))
		time.Sleep(1 * time.Second)
	}
}
