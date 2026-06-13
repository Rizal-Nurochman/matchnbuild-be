package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait       = 10 * time.Second
	pongWait        = 60 * time.Second
	pingPeriod      = (pongWait * 9) / 10
	maxMessageSize  = 65536
	sendBufferSize  = 256
	maxEventsPerSec = 50
)

type Client struct {
	hub          *Hub
	conn         *websocket.Conn
	userID       string
	send         chan []byte
	rooms        map[string]bool
	mu           sync.RWMutex
	rateLimiter  *RateLimiter
	closeOnce    sync.Once
}

type RateLimiter struct {
	events    []time.Time
	mu        sync.Mutex
	maxEvents int
	window    time.Duration
}

func NewRateLimiter(maxEvents int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		maxEvents: maxEvents,
		window:    window,
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	var validEvents []time.Time
	for _, t := range rl.events {
		if t.After(cutoff) {
			validEvents = append(validEvents, t)
		}
	}

	if len(validEvents) >= rl.maxEvents {
		rl.events = validEvents
		return false
	}

	rl.events = append(validEvents, now)
	return true
}

type IncomingMessage struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Data      json.RawMessage `json:"data"`
}

func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:         hub,
		conn:        conn,
		userID:      userID,
		send:        make(chan []byte, sendBufferSize),
		rooms:       make(map[string]bool),
		rateLimiter: NewRateLimiter(maxEventsPerSec, time.Second),
	}
}

func (c *Client) JoinRoom(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rooms[roomID] = true
}

func (c *Client) LeaveRoom(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.rooms, roomID)
}

func (c *Client) IsInRoom(roomID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rooms[roomID]
}

func (c *Client) UserID() string {
	return c.userID
}

func (c *Client) ReadPump(handler func(client *Client, msg IncomingMessage)) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] read error: %v", err)
			}
			break
		}

		if !c.rateLimiter.Allow() {
			continue
		}

		var msg IncomingMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		handler(c, msg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.send)
	})
}
