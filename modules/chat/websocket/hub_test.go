package websocket

import (
	"sync"
	"testing"
	"time"
)

func newTestClient(hub *Hub, userID string, rooms []string) *Client {
	return &Client{
		hub:          hub,
		userID:       userID,
		send:         make(chan []byte, sendBufferSize),
		rooms:        make(map[string]bool),
		pendingRooms: rooms,
		rateLimiter:  NewRateLimiter(maxEventsPerSec, time.Second),
	}
}

func drain(c *Client, wg *sync.WaitGroup) {
	defer wg.Done()
	for range c.send {
	}
}

func TestHubConcurrency(t *testing.T) {
	var presenceMu sync.Mutex
	presenceCalls := 0

	hub := NewHub(func(userID string, isOnline bool, lastSeenAt *time.Time, roomIDs []string) {
		presenceMu.Lock()
		presenceCalls++
		presenceMu.Unlock()
	})
	go hub.Run()
	defer hub.Shutdown()

	const (
		workers   = 40
		rounds    = 25
		roomCount = 8
	)

	roomID := func(i int) string {
		return "room-" + string(rune('A'+i%roomCount))
	}

	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			userID := "user-" + string(rune('a'+w%26))
			for r := 0; r < rounds; r++ {
				rooms := []string{roomID(w), roomID(w + 1)}
				c := newTestClient(hub, userID, rooms)

				var dwg sync.WaitGroup
				dwg.Add(1)
				go drain(c, &dwg)

				hub.register <- c
				hub.BroadcastToRoomExcept(rooms[0], "message.created", map[string]string{"x": "y"}, "")
				hub.SendToUserEvent(userID, "ping", map[string]string{"a": "b"})
				hub.unregister <- c

				dwg.Wait()
			}
		}(w)
	}

	for r := 0; r < workers/2; r++ {
		wg.Add(1)
		go func(r int) {
			defer wg.Done()
			for i := 0; i < rounds*4; i++ {
				_ = hub.ActiveConnections()
				_ = hub.UserConnectionCount("user-a")
				_ = hub.IsUserOnline("user-b")
				hub.BroadcastToRooms([]string{roomID(r), roomID(r + 2)}, "presence.changed", map[string]bool{"online": true}, "user-c")
				hub.BroadcastToAllUsers("global", map[string]int{"n": i})
			}
		}(r)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("hub concurrency test timed out (possible deadlock)")
	}

	if hub.ActiveConnections() != 0 {
		t.Fatalf("expected 0 active connections after churn, got %d", hub.ActiveConnections())
	}
}

func TestHubSlowConsumerEviction(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	defer hub.Shutdown()

	slow := newTestClient(hub, "slow-user", []string{"room-X"})
	hub.register <- slow

	for i := 0; i < sendBufferSize+50; i++ {
		hub.BroadcastToRoomExcept("room-X", "message.created", map[string]int{"i": i}, "")
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if hub.ActiveConnections() == 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if hub.ActiveConnections() != 0 {
		t.Fatalf("expected slow consumer to be evicted, active=%d", hub.ActiveConnections())
	}
	if hub.UserConnectionCount("slow-user") != 0 {
		t.Fatalf("slow consumer still indexed after eviction")
	}
}
