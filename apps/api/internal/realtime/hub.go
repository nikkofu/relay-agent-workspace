package realtime

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Event struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	WorkspaceID string      `json:"workspace_id,omitempty"`
	ChannelID   string      `json:"channel_id,omitempty"`
	EntityID    string      `json:"entity_id,omitempty"`
	TS          string      `json:"ts"`
	Payload     interface{} `json:"payload"`
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	register   chan *client
	unregister chan *client
	broadcast  chan []byte

	mu      sync.RWMutex
	clients map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan []byte, 32),
		clients:    make(map[*client]struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = struct{}{}
			h.mu.Unlock()
		case c := <-h.unregister:
			h.removeClient(c)
		case message := <-h.broadcast:
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.send <- message:
				default:
					go h.removeClient(c)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Broadcast(event Event) error {
	data, err := marshalEvent(event)
	if err != nil {
		return err
	}

	h.broadcast <- data
	return nil
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) error {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	c := &client{
		conn: conn,
		send: make(chan []byte, 16),
	}

	h.register <- c

	connectedEvent, err := marshalEvent(Event{
		ID:      "evt_" + time.Now().Format("20060102150405.000000"),
		Type:    "realtime.connected",
		TS:      time.Now().UTC().Format(time.RFC3339Nano),
		Payload: map[string]any{"connected": true},
	})
	if err != nil {
		log.Printf("failed to broadcast websocket connect event: %v", err)
	} else {
		c.send <- connectedEvent
	}

	go h.writePump(c)
	h.readPump(c)

	return nil
}

func (h *Hub) readPump(c *client) {
	defer func() {
		h.unregister <- c
	}()

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *Hub) writePump(c *client) {
	defer func() {
		_ = c.conn.Close()
	}()

	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

func (h *Hub) removeClient(c *client) {
	h.mu.Lock()
	_, ok := h.clients[c]
	if ok {
		delete(h.clients, c)
		close(c.send)
	}
	h.mu.Unlock()

	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func marshalEvent(event Event) ([]byte, error) {
	return json.Marshal(event)
}
