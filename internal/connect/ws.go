package connect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AleksandrMac/testfsd/internal/datastores"
	"github.com/AleksandrMac/testfsd/internal/entities"
	"github.com/AleksandrMac/testfsd/internal/log"
	"github.com/AleksandrMac/testfsd/internal/queue"
	"github.com/AleksandrMac/testfsd/internal/services"
	"github.com/AleksandrMac/testfsd/pkg/ws"

	"github.com/gorilla/websocket"
)

type connector struct {
	ms services.MessageService
}

// NewConnector get Connect service instance
func NewConnector(q queue.Queuer, ds datastores.CommonDatastore) Connector {
	return &connector{ms: services.NewMessageService(q, ds)}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Пропускаем любой запрос
	},
}

// WS godoc
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
func (x *connector) WS(rw http.ResponseWriter, r *http.Request) {
	connection, _ := upgrader.Upgrade(rw, r, nil)
	defer connection.Close() // Закрываем соединение
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	var userId int64

	conn := ws.NewWS(connection)

	for {
		mt, message, err := connection.ReadMessage()
		if err != nil || mt == websocket.CloseMessage {
			log.Default().Info(fmt.Sprintf("user %d closed", userId))
			// x.ms.UnregisterWSClient(connection)
			break // Выходим из цикла, если клиент пытается закрыть соединение или связь с клиентом прервана
		}

		var method entities.Method

		if err = json.Unmarshal(message, &method); err != nil {
			log.Default().Warn("failed parse ws message: " + err.Error())
			conn.SendMessage([]byte(`{"error":"bad message"}`))
			// connection.WriteMessage(websocket.TextMessage, )
			break
		}

		switch {
		case method.Authorization != nil:
			if userId > 0 {
				continue
			}

			userId = method.Authorization.UserId
			// if err = x.ms.RegisterWSClient(method.Authorization.UserId, connection); err != nil {
			// 	log.Default().Warn("failed register ws client: " + err.Error())
			// 	connection.WriteMessage(websocket.TextMessage, []byte(`{"error":"unauthorized"}`))
			// 	break
			// }
			listenRooms, err := x.ms.GetRooms(ctx, method.Authorization.UserId)
			if err != nil {
				log.Default().Warn("failed register ws client: " + err.Error())
				conn.SendMessage([]byte(`{"error":"unauthorized"}`))
				break
			}

			go x.ms.MessageSending(ctx, listenRooms, conn)
			if err = x.ms.Notification(ctx, &entities.Notification{
				Event: "connected",
				Metadata: map[string]interface{}{
					"user": entities.User{Id: userId},
				}}, conn); err != nil {
				log.Default().Warn("failed push notification")
			}
			message = []byte(`{"status": "ok"}`)
		case method.ReadMessageList != nil:
			messages, err := x.ms.GetMessages(ctx, method.ReadMessageList.RoomId, method.ReadMessageList.AfterMessageId)
			if err != nil {
				log.Default().Warn("failed read message list: " + err.Error())
				connection.WriteMessage(websocket.TextMessage, []byte(`{"error":"failed read message list"}`))
				continue
			}
			if message, err = json.Marshal(messages); err != nil {
				log.Default().Warn("failed marshal message list: " + err.Error())
				connection.WriteMessage(websocket.TextMessage, []byte(`{"error":"failed read message list"}`))
				continue
			}
		case method.WriteMessage != nil:
			if userId < 1 {
				conn.SendMessage([]byte(`{"error":"unauthorized"}`))
				continue
			}
			msg := method.WriteMessage
			msg.Author.Id = userId
			id, err := x.ms.WriteMessage(ctx, msg, conn)
			if err != nil {
				log.Default().Warn("failed write message: " + err.Error())
				conn.SendMessage([]byte(`{"error":"failed write message"}`))
				continue
			}
			message = []byte(fmt.Sprintf(`{"message": {"id":%d}}`, id))
		case method.Rooms != nil:
			if userId < 1 {
				conn.SendMessage([]byte(`{"error":"unauthorized"}`))
				continue
			}
			rooms, err := x.ms.GetRooms(ctx, userId)
			if err != nil {
				log.Default().Warn("failed read message list: " + err.Error())
				conn.SendMessage([]byte(`{"error":"failed read message list"}`))
				continue
			}
			if message, err = json.Marshal(rooms); err != nil {
				log.Default().Warn("failed marshal message list: " + err.Error())
				conn.SendMessage([]byte(`{"error":"failed read message list"}`))
				continue
			}
		default:
			log.Default().Debug("unknown method")
			conn.SendMessage([]byte("unknown method"))
			break
		}

		conn.SendMessage(message)
	}
}
