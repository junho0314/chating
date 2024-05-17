package controller

import (
	"net/http"
	"time"

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
	connections map[*websocket.Conn]chan []byte
	broadcast   chan []byte
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		broadcast:   make(chan []byte),
		register:    make(chan *websocket.Conn),
		unregister:  make(chan *websocket.Conn),
		connections: make(map[*websocket.Conn]chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.connections[conn] = make(chan []byte)
			go h.writePump(conn)
		case conn := <-h.unregister:
			if _, ok := h.connections[conn]; ok {
				delete(h.connections, conn)
				close(h.connections[conn])

			}
		case message := <-h.broadcast:
			for conn := range h.connections {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Err(err).Msgf("Failed to write message : %v", err)
					conn.Close()
					delete(h.connections, conn)
				}
			}
		}
	}
}

func (h *Hub) writePump(conn *websocket.Conn) {
	ticker := time.NewTicker(54 * time.Second) // Ping-Pong을 위한 주기적 타이머
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-h.connections[conn]:
			if !ok {
				// Hub가 채널을 닫음
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Err(err).Msg("Failed to write message")
				return
			}
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Err(err).Msg("Failed to send ping message")
				return
			}
		}
	}
}

func WebsocketHandler(hub *Hub, w http.ResponseWriter, r *http.Request, roomId string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Err(err).Msgf("Failed to upgrade connection : %v", err)
		return
	}
	hub.register <- conn

	defer func() {
		hub.unregister <- conn
	}()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Err(err).Msgf("Failed to read message : %v", err)
			break
		}
		hub.broadcast <- message
	}
}
