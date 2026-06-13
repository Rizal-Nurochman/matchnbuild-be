package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

type Hub struct {
	clients    map[*Client]bool
	userIndex  map[string]map[*Client]bool
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	joinRoom   chan *ClientRoom
	leaveRoom  chan *ClientRoom
	broadcast  chan *RoomMessage
	userMsg    chan *UserMessage
	mu         sync.RWMutex
	onPresence func(userID string, isOnline bool, lastSeenAt *time.Time)
}

func (h *Hub) SetOnPresence(fn func(userID string, isOnline bool, lastSeenAt *time.Time)) {
	h.onPresence = fn
}

type ClientRoom struct {
	Client *Client
	RoomID string
}

type RoomMessage struct {
	RoomID  string
	Payload []byte
	Exclude string
}

type UserMessage struct {
	UserID  string
	Payload []byte
}

func NewHub(onPresence func(userID string, isOnline bool, lastSeenAt *time.Time)) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		userIndex:  make(map[string]map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		joinRoom:   make(chan *ClientRoom),
		leaveRoom:  make(chan *ClientRoom),
		broadcast:  make(chan *RoomMessage, 256),
		userMsg:    make(chan *UserMessage, 256),
		onPresence: onPresence,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			wasOnline := h.userIndex[client.userID] != nil && len(h.userIndex[client.userID]) > 0
			h.clients[client] = true
			if h.userIndex[client.userID] == nil {
				h.userIndex[client.userID] = make(map[*Client]bool)
			}
			h.userIndex[client.userID][client] = true
			nowOnline := len(h.userIndex[client.userID]) > 0
			h.mu.Unlock()
			log.Printf("[WS] client connected: user=%s total=%d", client.userID, len(h.clients))

			if !wasOnline && nowOnline && h.onPresence != nil {
				h.onPresence(client.userID, true, nil)
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if conns, ok := h.userIndex[client.userID]; ok {
					delete(conns, client)
					nowOnline := len(conns) > 0
					if len(conns) == 0 {
						delete(h.userIndex, client.userID)
					}
					if !nowOnline && h.onPresence != nil {
						now := time.Now()
						h.onPresence(client.userID, false, &now)
					}
				}
				for roomID := range client.rooms {
					if room, ok := h.rooms[roomID]; ok {
						delete(room, client)
						if len(room) == 0 {
							delete(h.rooms, roomID)
						}
					}
				}
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[WS] client disconnected: user=%s total=%d", client.userID, len(h.clients))

		case cr := <-h.joinRoom:
			h.mu.Lock()
			if h.rooms[cr.RoomID] == nil {
				h.rooms[cr.RoomID] = make(map[*Client]bool)
			}
			h.rooms[cr.RoomID][cr.Client] = true
			cr.Client.JoinRoom(cr.RoomID)
			h.mu.Unlock()
			log.Printf("[WS] client joined room: user=%s room=%s", cr.Client.userID, cr.RoomID)

		case cr := <-h.leaveRoom:
			h.mu.Lock()
			if room, ok := h.rooms[cr.RoomID]; ok {
				delete(room, cr.Client)
				if len(room) == 0 {
					delete(h.rooms, cr.RoomID)
				}
			}
			cr.Client.LeaveRoom(cr.RoomID)
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			room, ok := h.rooms[msg.RoomID]
			if ok {
				for client := range room {
					if client.userID == msg.Exclude {
						continue
					}
					select {
					case client.send <- msg.Payload:
					default:
						close(client.send)
						delete(room, client)
					}
				}
			}
			h.mu.RUnlock()

		case msg := <-h.userMsg:
			h.mu.RLock()
			if conns, ok := h.userIndex[msg.UserID]; ok {
				for client := range conns {
					select {
					case client.send <- msg.Payload:
					default:
						close(client.send)
						delete(conns, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToRoom(roomID string, payload []byte, excludeUserID string) {
	h.broadcast <- &RoomMessage{
		RoomID:  roomID,
		Payload: payload,
		Exclude: excludeUserID,
	}
}

func (h *Hub) SendToUser(userID string, payload []byte) {
	h.userMsg <- &UserMessage{
		UserID:  userID,
		Payload: payload,
	}
}

func (h *Hub) BroadcastToRoomExcept(roomID string, eventType string, data interface{}, excludeUserID string) {
	resp := map[string]interface{}{
		"type": eventType,
		"data": data,
	}
	payload, err := json.Marshal(resp)
	if err != nil {
		return
	}
	h.BroadcastToRoom(roomID, payload, excludeUserID)
}

func (h *Hub) SendToUserEvent(userID string, eventType string, data interface{}) {
	resp := map[string]interface{}{
		"type": eventType,
		"data": data,
	}
	payload, err := json.Marshal(resp)
	if err != nil {
		return
	}
	h.SendToUser(userID, payload)
}

func (h *Hub) UserConnectionCount(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userIndex[userID])
}

func (h *Hub) IsUserOnline(userID string) bool {
	return h.UserConnectionCount(userID) > 0
}

func (h *Hub) JoinUserToRoom(userID string, roomID string) {
	h.mu.RLock()
	conns := h.userIndex[userID]
	h.mu.RUnlock()

	for client := range conns {
		h.joinRoom <- &ClientRoom{
			Client: client,
			RoomID: roomID,
		}
	}
}

func (h *Hub) BroadcastToAllUsers(eventType string, data interface{}) {
	resp := map[string]interface{}{
		"type": eventType,
		"data": data,
	}
	payload, err := json.Marshal(resp)
	if err != nil {
		return
	}

	h.mu.RLock()
	for _, conns := range h.userIndex {
		for client := range conns {
			select {
			case client.send <- payload:
			default:
			}
		}
	}
	h.mu.RUnlock()
}
