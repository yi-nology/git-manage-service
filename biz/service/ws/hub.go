package ws

import (
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID     string
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	Topics map[string]bool
	connMu sync.Mutex
	closed atomic.Bool
}

type Message struct {
	Topic   string      `json:"topic"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

var DefaultHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WS] Client connected: %s (total: %d)", client.ID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			var msg Message
			if json.Unmarshal(message, &msg) == nil {
				for client := range h.clients {
					if client.Topics != nil {
						if !client.Topics[msg.Topic] && !client.Topics["*"] {
							continue
						}
					}
					select {
					case client.Send <- message:
					default:
						go func(c *Client) {
							h.unregister <- c
						}(client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	if client.closed.CompareAndSwap(false, true) {
		h.unregister <- client
	}
}

func (h *Hub) Broadcast(topic, msgType string, payload interface{}) {
	data, err := json.Marshal(Message{
		Topic:   topic,
		Type:    msgType,
		Payload: payload,
	})
	if err != nil {
		return
	}
	select {
	case h.broadcast <- data:
	default:
		log.Printf("[WS] Broadcast channel full, dropping message for topic %s", topic)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		c.Hub.Unregister(c)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.connMu.Lock()
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.connMu.Unlock()
				return
			}
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.connMu.Unlock()
				return
			}
			c.connMu.Unlock()

		case <-ticker.C:
			c.connMu.Lock()
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.connMu.Unlock()
				return
			}
			c.connMu.Unlock()
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] Read error: %v", err)
			}
			break
		}

		var cmd struct {
			Action string   `json:"action"`
			Topics []string `json:"topics"`
		}
		if json.Unmarshal(message, &cmd) == nil {
			switch cmd.Action {
			case "subscribe":
				if c.Topics == nil {
					c.Topics = make(map[string]bool)
				}
				for _, t := range cmd.Topics {
					c.Topics[t] = true
				}
			case "unsubscribe":
				for _, t := range cmd.Topics {
					delete(c.Topics, t)
				}
			}
		}
	}
}
