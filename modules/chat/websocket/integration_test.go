package websocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	authService "github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	chatDto "github.com/Rizal-Nurochman/matchnbuild/modules/chat/dto"
	chatRepo "github.com/Rizal-Nurochman/matchnbuild/modules/chat/repository"
	"github.com/Rizal-Nurochman/matchnbuild/modules/chat/service"
	prRepo "github.com/Rizal-Nurochman/matchnbuild/modules/project_request/repository"
	userRepository "github.com/Rizal-Nurochman/matchnbuild/modules/user/repository"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var dbCounter int64

type seedIDs struct {
	clientID       string
	designerID     string
	outsiderID     string
	conversationID string
	clientToken    string
	designerToken  string
	outsiderToken  string
}

type wsEvent struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id"`
	Data      json.RawMessage `json:"data"`
}

type wsClient struct {
	conn    *websocket.Conn
	label   string
	pending []wsEvent
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:chatmem%d?mode=memory&cache=shared", atomic.AddInt64(&dbCounter, 1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		TranslateError:                           true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	stmts := []string{
		`CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT,
			email TEXT,
			password TEXT,
			role TEXT,
			profile_picture TEXT,
			is_verified NUMERIC,
			verification_code TEXT,
			verification_expiry DATETIME,
			last_seen_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE conversations (
			id TEXT PRIMARY KEY,
			project_request_id TEXT,
			order_id TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE conversation_participants (
			conversation_id TEXT,
			user_id TEXT,
			role TEXT,
			last_read_message_id TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			PRIMARY KEY (conversation_id, user_id)
		)`,
		`CREATE TABLE messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT,
			sender_id TEXT,
			client_message_id TEXT,
			message_text TEXT,
			attachment_url TEXT,
			message_type TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			UNIQUE (sender_id, client_message_id)
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create schema: %v", err)
		}
	}

	t.Cleanup(func() {
		time.Sleep(50 * time.Millisecond)
		sqlDB.Close()
	})
	return db
}

func seedChat(t *testing.T, db *gorm.DB) seedIDs {
	t.Helper()

	client := entities.User{ID: uuid.New(), Name: "Client", Email: uuid.NewString() + "@test.com", Password: "x", Role: constants.ENUM_ROLE_CLIENT}
	designer := entities.User{ID: uuid.New(), Name: "Designer", Email: uuid.NewString() + "@test.com", Password: "x", Role: constants.ENUM_ROLE_DESIGNER}
	outsider := entities.User{ID: uuid.New(), Name: "Outsider", Email: uuid.NewString() + "@test.com", Password: "x", Role: constants.ENUM_ROLE_CLIENT}

	if err := db.Create(&[]entities.User{client, designer, outsider}).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}

	conv := entities.Conversation{ID: uuid.New(), ProjectRequestID: uuid.New()}
	if err := db.Create(&conv).Error; err != nil {
		t.Fatalf("seed conversation: %v", err)
	}

	participants := []entities.ConversationParticipant{
		{ConversationID: conv.ID, UserID: client.ID, Role: constants.ENUM_ROLE_CLIENT},
		{ConversationID: conv.ID, UserID: designer.ID, Role: constants.ENUM_ROLE_DESIGNER},
	}
	if err := db.Create(&participants).Error; err != nil {
		t.Fatalf("seed participants: %v", err)
	}

	return seedIDs{
		clientID:       client.ID.String(),
		designerID:     designer.ID.String(),
		outsiderID:     outsider.ID.String(),
		conversationID: conv.ID.String(),
	}
}

func setupTestServer(t *testing.T) (string, *gorm.DB, *Hub, seedIDs) {
	t.Helper()

	db := newTestDB(t)
	ids := seedChat(t, db)

	messageRepo := chatRepo.NewMessageRepository(db)
	convRepo := prRepo.NewConversationRepository(db)
	participantRepo := prRepo.NewConversationParticipantRepository(db)
	userRepo := userRepository.NewUserRepository(db)
	svc := service.NewChatService(messageRepo, participantRepo, convRepo, db)

	hub := NewHub(nil)
	go hub.Run()
	svc.SetBroadcaster(hub)

	jwtSvc := authService.NewJWTService()
	handler := NewHandler(hub, svc, userRepo, jwtSvc, db, "", false)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/api/v1/ws", handler.HandleWebSocket)

	srv := httptest.NewServer(engine)
	t.Cleanup(func() {
		srv.Close()
		hub.Shutdown()
	})

	ids.clientToken = jwtSvc.GenerateAccessToken(ids.clientID, constants.ENUM_ROLE_CLIENT)
	ids.designerToken = jwtSvc.GenerateAccessToken(ids.designerID, constants.ENUM_ROLE_DESIGNER)
	ids.outsiderToken = jwtSvc.GenerateAccessToken(ids.outsiderID, constants.ENUM_ROLE_CLIENT)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v1/ws"
	return wsURL, db, hub, ids
}

func dial(t *testing.T, wsURL, label, token string) *wsClient {
	t.Helper()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL+"?token="+token, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return &wsClient{conn: conn, label: label}
}

func waitOnline(t *testing.T, hub *Hub, userID string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hub.IsUserOnline(userID) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("user %s did not come online in time", userID)
}

func (c *wsClient) send(t *testing.T, eventType, requestID string, data any) {
	t.Helper()

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}

	envelope := map[string]any{"type": eventType, "data": json.RawMessage(raw)}
	if requestID != "" {
		envelope["request_id"] = requestID
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		t.Fatalf("write message: %v", err)
	}

	t.Logf(">> [%s] SEND %s %s", c.label, eventType, string(raw))
}

func (c *wsClient) read(t *testing.T, wantType string, timeout time.Duration) wsEvent {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		for i, ev := range c.pending {
			if ev.Type == wantType {
				c.pending = append(c.pending[:i], c.pending[i+1:]...)
				return ev
			}
		}

		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for event %q (buffered: %v)", wantType, eventTypes(c.pending))
		}

		c.conn.SetReadDeadline(deadline)
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			t.Fatalf("read while waiting for %q: %v", wantType, err)
		}

		for _, line := range bytes.Split(raw, []byte("\n")) {
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			var ev wsEvent
			if err := json.Unmarshal(line, &ev); err != nil {
				continue
			}
			t.Logf("<< [%s] RECV %s %s", c.label, ev.Type, string(ev.Data))
			c.pending = append(c.pending, ev)
		}
	}
}

func eventTypes(events []wsEvent) []string {
	types := make([]string, len(events))
	for i, e := range events {
		types[i] = e.Type
	}
	return types
}

func messageCount(t *testing.T, db *gorm.DB, query string, args ...any) int64 {
	t.Helper()

	var count int64
	if err := db.Model(&entities.Message{}).Where(query, args...).Count(&count).Error; err != nil {
		t.Fatalf("count messages: %v", err)
	}
	return count
}

func TestChatIntegration(t *testing.T) {
	t.Run("delivery: client message reaches designer and is persisted", func(t *testing.T) {
		wsURL, db, hub, ids := setupTestServer(t)

		client := dial(t, wsURL, "client", ids.clientToken)
		waitOnline(t, hub, ids.clientID)
		designer := dial(t, wsURL, "designer", ids.designerToken)
		waitOnline(t, hub, ids.designerID)

		reqID := "req-delivery"
		clientMsgID := uuid.NewString()
		client.send(t, "message.send", reqID, map[string]any{
			"conversation_id":   ids.conversationID,
			"client_message_id": clientMsgID,
			"text":              "halo designer",
			"message_type":      "Text",
		})

		ack := client.read(t, "message.created", 2*time.Second)
		if ack.RequestID != reqID {
			t.Errorf("ack request_id = %q, want %q", ack.RequestID, reqID)
		}
		var ackData chatDto.MessageCreatedData
		if err := json.Unmarshal(ack.Data, &ackData); err != nil {
			t.Fatalf("unmarshal ack: %v", err)
		}
		if ackData.SenderID != ids.clientID {
			t.Errorf("ack sender_id = %q, want %q", ackData.SenderID, ids.clientID)
		}
		if ackData.Text != "halo designer" {
			t.Errorf("ack text = %q, want %q", ackData.Text, "halo designer")
		}

		peer := designer.read(t, "message.created", 2*time.Second)
		var peerData chatDto.MessageCreatedData
		if err := json.Unmarshal(peer.Data, &peerData); err != nil {
			t.Fatalf("unmarshal peer: %v", err)
		}
		if peerData.SenderID != ids.clientID {
			t.Errorf("peer sender_id = %q, want %q", peerData.SenderID, ids.clientID)
		}
		if peerData.Text != "halo designer" {
			t.Errorf("peer text = %q, want %q", peerData.Text, "halo designer")
		}

		if got := messageCount(t, db, "conversation_id = ?", ids.conversationID); got != 1 {
			t.Errorf("message count = %d, want 1", got)
		}
	})

	t.Run("idempotency: duplicate client_message_id persists once", func(t *testing.T) {
		wsURL, db, hub, ids := setupTestServer(t)

		client := dial(t, wsURL, "client", ids.clientToken)
		waitOnline(t, hub, ids.clientID)

		clientMsgID := uuid.NewString()
		payload := map[string]any{
			"conversation_id":   ids.conversationID,
			"client_message_id": clientMsgID,
			"text":              "pesan kembar",
			"message_type":      "Text",
		}

		client.send(t, "message.send", "req-1", payload)
		first := client.read(t, "message.created", 2*time.Second)

		client.send(t, "message.send", "req-2", payload)
		second := client.read(t, "message.created", 2*time.Second)

		var firstData, secondData chatDto.MessageCreatedData
		if err := json.Unmarshal(first.Data, &firstData); err != nil {
			t.Fatalf("unmarshal first: %v", err)
		}
		if err := json.Unmarshal(second.Data, &secondData); err != nil {
			t.Fatalf("unmarshal second: %v", err)
		}
		if firstData.ID != secondData.ID {
			t.Errorf("duplicate send returned different message ids: %q vs %q", firstData.ID, secondData.ID)
		}

		if got := messageCount(t, db, "client_message_id = ?", clientMsgID); got != 1 {
			t.Errorf("message count for client_message_id = %d, want 1", got)
		}
	})

	t.Run("authorization: non-participant is rejected", func(t *testing.T) {
		wsURL, db, hub, ids := setupTestServer(t)

		outsider := dial(t, wsURL, "outsider", ids.outsiderToken)
		waitOnline(t, hub, ids.outsiderID)

		outsider.send(t, "message.send", "req-intrusion", map[string]any{
			"conversation_id":   ids.conversationID,
			"client_message_id": uuid.NewString(),
			"text":              "aku menyusup",
			"message_type":      "Text",
		})

		ev := outsider.read(t, "error", 2*time.Second)
		var errData chatDto.WSErrorData
		if err := json.Unmarshal(ev.Data, &errData); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if errData.Code != "FORBIDDEN" {
			t.Errorf("error code = %q, want FORBIDDEN", errData.Code)
		}

		if got := messageCount(t, db, "conversation_id = ?", ids.conversationID); got != 0 {
			t.Errorf("message count = %d, want 0 (nothing should be persisted)", got)
		}
	})
}
