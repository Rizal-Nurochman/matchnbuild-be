package websocket

import (
	"sync"
	"testing"
	"time"
)

// newTestClient builds a Client without a real websocket connection. The hub
// never touches Client.conn (only Client.send), so a nil conn is safe for
// hub-level concurrency tests.
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

// drain keeps reading a client's send channel until it is closed so buffered
// sends never block the hub. Exits cleanly when Close() closes the channel.
func drain(c *Client, wg *sync.WaitGroup) {
	defer wg.Done()
	for range c.send {
	}
}

// TestHubConcurrency hammers the hub with concurrent register/unregister,
// room broadcasts, user messages, presence-scoped broadcasts and read-only
// queries. Run with -race to catch concurrent map access or send-on-closed
// channel panics (the C2/C3 regressions this guards against).
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

	// Connection churn: each worker repeatedly connects and disconnects a client
	// that belongs to a couple of rooms.
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

	// Concurrent readers that take RLock while the Run loop mutates maps under Lock.
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

// TestHubSlowConsumerEviction verifies that a client whose send buffer is full
// is evicted during a broadcast without panicking and without leaking from the
// hub indexes.
func TestHubSlowConsumerEviction(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	defer hub.Shutdown()

	// A client we never drain: its send buffer will fill up and it should be
	// evicted on the next broadcast once the buffer is exhausted.
	slow := newTestClient(hub, "slow-user", []string{"room-X"})
	hub.register <- slow

	// Fill the send buffer beyond capacity via repeated room broadcasts.
	for i := 0; i < sendBufferSize+50; i++ {
		hub.BroadcastToRoomExcept("room-X", "message.created", map[string]int{"i": i}, "")
	}

	// Give the Run loop time to process the broadcast backlog and evict.
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
