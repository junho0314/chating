package controller

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS 문제를 방지하기 위해 모든 오리진에서의 웹소켓 요청을 허용합니다.
	},
}

type Hub struct {
	rooms      map[string]map[*websocket.Conn]chan []byte
	broadcast  chan Message
	register   chan Subscription
	unregister chan Subscription
	mu         sync.Mutex
}
type Message struct {
	roomId string
	data   []byte
}

type Subscription struct {
	conn   *websocket.Conn
	roomId string
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan Subscription),
		unregister: make(chan Subscription),
		rooms:      make(map[string]map[*websocket.Conn]chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case subscription := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[subscription.roomId]; !ok {
				h.rooms[subscription.roomId] = make(map[*websocket.Conn]chan []byte)
			}
			h.rooms[subscription.roomId][subscription.conn] = make(chan []byte)
			go h.writePump(subscription.conn, subscription.roomId)
			go h.readPump(subscription.conn, subscription.roomId)
			h.mu.Unlock()
		case subscription := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.rooms[subscription.roomId]; ok {
				if _, ok := h.rooms[subscription.roomId][subscription.conn]; ok {
					close(h.rooms[subscription.roomId][subscription.conn])
					delete(h.rooms[subscription.roomId], subscription.conn)
					if len(h.rooms[subscription.roomId]) == 0 {
						delete(h.rooms, subscription.roomId)
					}
				}

			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for conn, send := range h.rooms[message.roomId] {
				select {
				case send <- message.data:
				default:
					close(send)
					delete(h.rooms[message.roomId], conn)
					conn.Close()
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) writePump(conn *websocket.Conn, roomId string) {
	for {
		select {
		case message, ok := <-h.rooms[roomId][conn]:
			if !ok {
				// Hub가 채널을 닫음
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Err(err).Msg("Failed to write message")
				return
			}
		}
	}
}

func (h *Hub) readPump(conn *websocket.Conn, roomId string) {
	defer func() {
		h.unregister <- Subscription{conn: conn, roomId: roomId}
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Err(err).Msg("Unexpected close error")
			} else {
				log.Err(err).Msg("Failed to read message")
			}
			break
		}
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			if msg["type"] == "ping" {
				log.Info().Msg("Received ping")
				continue
			}
		}
		h.broadcast <- Message{roomId: roomId, data: message}
	}
}

func WebsocketHandler(hub *Hub, ginCtx *gin.Context) {
	roomId := ginCtx.Param("roomId")
	if roomId == "" {
		log.Warn().Msg("roomId is required")
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": "roomId is required"})
		return
	}

	// tokenString := ginCtx.Query("token")
	// if tokenString == "" {
	// 	ginCtx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
	// 	return
	// }

	// // JWT 토큰 검증
	// token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	// 	// 검증 키를 반환
	// 	return []byte("your_secret_key"), nil
	// })

	// if err != nil || !token.Valid {
	// 	ginCtx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
	// 	return
	// }

	conn, err := upgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)
	if err != nil {
		log.Err(err).Msgf("Failed to upgrade connection : %v", err)
		return
	}
	hub.register <- Subscription{conn: conn, roomId: roomId}
}
