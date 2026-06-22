package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
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
	shutdown   chan struct{}
	mu         sync.RWMutex
	onPresence func(userID string, isOnline bool, lastSeenAt *time.Time, roomIDs []string)

	// metrics
	totalConnections   int64
	activeConnections  int64
	totalMessages      int64
	totalErrors        int64
}

func (h *Hub) SetOnPresence(fn func(userID string, isOnline bool, lastSeenAt *time.Time, roomIDs []string)) {
	h.onPresence = fn
}

type offlineEvent struct {
	userID  string
	roomIDs []string
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

func NewHub(onPresence func(userID string, isOnline bool, lastSeenAt *time.Time, roomIDs []string)) *Hub {
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
		shutdown:   make(chan struct{}),
		onPresence: onPresence,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.shutdown:
			h.shutdownAll()
			return

		case client := <-h.register:
			h.mu.Lock()
			wasOnline := h.userIndex[client.userID] != nil && len(h.userIndex[client.userID]) > 0
			h.clients[client] = true
			if h.userIndex[client.userID] == nil {
				h.userIndex[client.userID] = make(map[*Client]bool)
			}
			h.userIndex[client.userID][client] = true
			nowOnline := len(h.userIndex[client.userID]) > 0
			roomIDs := make([]string, 0, len(client.pendingRooms))
			for _, roomID := range client.pendingRooms {
				if h.rooms[roomID] == nil {
					h.rooms[roomID] = make(map[*Client]bool)
				}
				h.rooms[roomID][client] = true
				client.rooms[roomID] = true
				roomIDs = append(roomIDs, roomID)
			}
			client.pendingRooms = nil
			atomic.AddInt64(&h.totalConnections, 1)
			atomic.AddInt64(&h.activeConnections, 1)
			h.mu.Unlock()
			log.Printf("[WS] client connected: user=%s total=%d", client.userID, h.ActiveConnections())

			if !wasOnline && nowOnline && h.onPresence != nil {
				h.onPresence(client.userID, true, nil, roomIDs)
			}

		case client := <-h.unregister:
			h.mu.Lock()
			ev := h.removeClientLocked(client)
			h.mu.Unlock()
			log.Printf("[WS] client disconnected: user=%s total=%d", client.userID, h.ActiveConnections())

			h.notifyOffline([]offlineEvent{ev})

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
			atomic.AddInt64(&h.totalMessages, 1)
			var dead []*Client
			h.mu.Lock()
			if room, ok := h.rooms[msg.RoomID]; ok {
				for client := range room {
					if client.userID == msg.Exclude {
						continue
					}
					select {
					case client.send <- msg.Payload:
					default:
						dead = append(dead, client)
					}
				}
			}
			offline := h.removeClientsLocked(dead)
			h.mu.Unlock()
			h.notifyOffline(offline)

		case msg := <-h.userMsg:
			atomic.AddInt64(&h.totalMessages, 1)
			var dead []*Client
			h.mu.Lock()
			if conns, ok := h.userIndex[msg.UserID]; ok {
				for client := range conns {
					select {
					case client.send <- msg.Payload:
					default:
						dead = append(dead, client)
					}
				}
			}
			offline := h.removeClientsLocked(dead)
			h.mu.Unlock()
			h.notifyOffline(offline)
		}
	}
}

func (h *Hub) removeClientLocked(client *Client) offlineEvent {
	if _, ok := h.clients[client]; !ok {
		return offlineEvent{}
	}
	delete(h.clients, client)
	atomic.AddInt64(&h.activeConnections, -1)

	wentOffline := false
	if conns, ok := h.userIndex[client.userID]; ok {
		delete(conns, client)
		if len(conns) == 0 {
			delete(h.userIndex, client.userID)
			wentOffline = true
		}
	}

	roomIDs := make([]string, 0, len(client.rooms))
	for roomID := range client.rooms {
		roomIDs = append(roomIDs, roomID)
		if room, ok := h.rooms[roomID]; ok {
			delete(room, client)
			if len(room) == 0 {
				delete(h.rooms, roomID)
			}
		}
	}
	client.Close()

	if !wentOffline {
		return offlineEvent{}
	}
	return offlineEvent{userID: client.userID, roomIDs: roomIDs}
}

func (h *Hub) removeClientsLocked(clients []*Client) []offlineEvent {
	var offline []offlineEvent
	for _, c := range clients {
		if ev := h.removeClientLocked(c); ev.userID != "" {
			offline = append(offline, ev)
		}
	}
	return offline
}

func (h *Hub) notifyOffline(events []offlineEvent) {
	if h.onPresence == nil {
		return
	}
	for _, ev := range events {
		if ev.userID == "" {
			continue
		}
		now := time.Now()
		h.onPresence(ev.userID, false, &now, ev.roomIDs)
	}
}

func (h *Hub) shutdownAll() {
	h.mu.Lock()
	for client := range h.clients {
		client.Close()
	}
	h.clients = nil
	h.userIndex = nil
	h.rooms = nil
	h.mu.Unlock()
	log.Println("[WS] Hub shutdown complete")
}

func (h *Hub) Shutdown() {
	close(h.shutdown)
}

func (h *Hub) ActiveConnections() int64 {
	return atomic.LoadInt64(&h.activeConnections)
}

func (h *Hub) TotalConnections() int64 {
	return atomic.LoadInt64(&h.totalConnections)
}

func (h *Hub) TotalMessages() int64 {
	return atomic.LoadInt64(&h.totalMessages)
}

func (h *Hub) TotalErrors() int64 {
	return atomic.LoadInt64(&h.totalErrors)
}

func (h *Hub) IncrementErrors() {
	atomic.AddInt64(&h.totalErrors, 1)
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

func (h *Hub) BroadcastToRooms(roomIDs []string, eventType string, data interface{}, excludeUserID string) {
	if len(roomIDs) == 0 {
		return
	}
	resp := map[string]interface{}{
		"type": eventType,
		"data": data,
	}
	payload, err := json.Marshal(resp)
	if err != nil {
		return
	}

	seen := make(map[*Client]bool)
	h.mu.RLock()
	for _, roomID := range roomIDs {
		room, ok := h.rooms[roomID]
		if !ok {
			continue
		}
		for client := range room {
			if client.userID == excludeUserID || seen[client] {
				continue
			}
			seen[client] = true
			select {
			case client.send <- payload:
			default:
			}
		}
	}
	h.mu.RUnlock()
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
	clients := make([]*Client, 0, len(conns))
	for client := range conns {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		h.joinRoom <- &ClientRoom{
			Client: client,
			RoomID: roomID,
		}
	}
}

func (h *Hub) JoinClientToRoom(client *Client, roomID string) {
	h.joinRoom <- &ClientRoom{
		Client: client,
		RoomID: roomID,
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
