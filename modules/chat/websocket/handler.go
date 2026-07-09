package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	chatDto "github.com/Rizal-Nurochman/matchnbuild/modules/chat/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/chat/service"
	"github.com/Rizal-Nurochman/matchnbuild/modules/user/repository"
	authService "github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type Handler struct {
	hub            *Hub
	chatSvc        service.ChatService
	userRepo       repository.UserRepository
	jwtService     authService.JWTService
	db             *gorm.DB
	allowedOrigins []string
	isProduction   bool
}

func NewHandler(hub *Hub, chatSvc service.ChatService, userRepo repository.UserRepository, jwtService authService.JWTService, db *gorm.DB, allowedOrigin string, isProduction bool) *Handler {
	h := &Handler{
		hub:            hub,
		chatSvc:        chatSvc,
		userRepo:       userRepo,
		jwtService:     jwtService,
		db:             db,
		allowedOrigins: parseOrigins(allowedOrigin),
		isProduction:   isProduction,
	}

	hub.SetOnPresence(h.handlePresenceChange)
	return h
}

func parseOrigins(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}

func (h *Handler) HandleWebSocket(ctx *gin.Context) {
	origin := ctx.Request.Header.Get("Origin")
	if !h.isOriginAllowed(origin) {
		log.Printf("[WS] origin ditolak: %q (allowlist=%v, prod=%v)", origin, h.allowedOrigins, h.isProduction)
		ctx.JSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
		h.hub.IncrementErrors()
		return
	}

	token := h.extractToken(ctx)
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		h.hub.IncrementErrors()
		return
	}

	userID, err := h.jwtService.GetUserIDByToken(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		h.hub.IncrementErrors()
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return h.isOriginAllowed(r.Header.Get("Origin"))
		},
	}

	var respHeader http.Header
	if proto := ctx.GetHeader("Sec-WebSocket-Protocol"); proto != "" {
		parts := strings.Split(proto, ",")
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "bearer" {
			respHeader = http.Header{"Sec-WebSocket-Protocol": []string{"bearer"}}
		}
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, respHeader)
	if err != nil {
		log.Printf("[WS] upgrade failed: %v", err)
		h.hub.IncrementErrors()
		return
	}

	client := NewClient(h.hub, conn, userID)

	if convs, cErr := h.chatSvc.GetConversations(ctx.Request.Context(), userID); cErr == nil {
		roomIDs := make([]string, 0, len(convs))
		for _, c := range convs {
			roomIDs = append(roomIDs, c.ID)
		}
		client.SetPendingRooms(roomIDs)
	} else {
		log.Printf("[WS] failed to preload conversations for user=%s: %v", userID, cErr)
	}

	h.hub.register <- client

	go client.WritePump()
	go client.ReadPump(h.handleEvent)
}

func (h *Handler) isOriginAllowed(origin string) bool {
	if origin == "" {
		return true
	}
	if len(h.allowedOrigins) == 0 {
		return !h.isProduction
	}
	for _, allowed := range h.allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func (h *Handler) extractToken(ctx *gin.Context) string {
	if auth := ctx.GetHeader("Authorization"); auth != "" {
		if after, ok := strings.CutPrefix(auth, "Bearer "); ok {
			return strings.TrimSpace(after)
		}
	}

	if proto := ctx.GetHeader("Sec-WebSocket-Protocol"); proto != "" {
		parts := strings.Split(proto, ",")
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "bearer" {
			return strings.TrimSpace(parts[1])
		}
	}

	return ctx.Query("token")
}

func (h *Handler) handleEvent(client *Client, msg IncomingMessage) {
	switch msg.Type {
	case chatDto.EVENT_MESSAGE_SEND:
		h.handleMessageSend(client, msg)
	case chatDto.EVENT_MESSAGE_READ:
		h.handleMessageRead(client, msg)
	case chatDto.EVENT_TYPING_START:
		h.handleTyping(client, msg, true)
	case chatDto.EVENT_TYPING_STOP:
		h.handleTyping(client, msg, false)
	case chatDto.EVENT_CONVERSATION_JOIN:
		h.handleConversationJoin(client, msg)
	default:
		h.sendError(client, msg.RequestID, "UNKNOWN_EVENT", "unknown event type")
	}
}

func (h *Handler) handleMessageSend(client *Client, msg IncomingMessage) {
	var data chatDto.MessageSendData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		h.sendError(client, msg.RequestID, "INVALID_PAYLOAD", "invalid message data")
		return
	}

	if data.ConversationID == "" || data.ClientMessageID == "" {
		h.sendError(client, msg.RequestID, "INVALID_PAYLOAD", "conversation_id and client_message_id required")
		return
	}

	if data.MessageType == "" {
		data.MessageType = "Text"
	}

	isParticipant, err := h.chatSvc.IsParticipant(client.UserID(), data.ConversationID)
	if err != nil {
		h.sendError(client, msg.RequestID, "SERVER_ERROR", "failed to verify participant")
		return
	}
	if !isParticipant {
		h.sendError(client, msg.RequestID, "FORBIDDEN", "you are not a participant of this conversation")
		return
	}

	resp, err := h.chatSvc.SendMessageWS(client.UserID(), data.ConversationID, chatDto.SendMessageRequest{
		ClientMessageID: data.ClientMessageID,
		MessageText:     data.Text,
		AttachmentURL:   data.AttachmentURL,
		MessageType:     data.MessageType,
	})
	if err != nil {
		h.sendError(client, msg.RequestID, "SEND_FAILED", err.Error())
		return
	}

	createdData := chatDto.MessageCreatedData{
		ID:              resp.ID,
		ConversationID:  resp.ConversationID,
		SenderID:        resp.SenderID,
		SenderName:      resp.SenderName,
		ClientMessageID: resp.ClientMessageID,
		MessageType:     resp.MessageType,
		Text:            resp.MessageText,
		AttachmentURL:   resp.AttachmentURL,
		CreatedAt:       resp.CreatedAt,
	}

	ackPayload, _ := json.Marshal(map[string]interface{}{
		"type":       chatDto.EVENT_MESSAGE_CREATED,
		"data":       createdData,
		"request_id": msg.RequestID,
	})
	select {
	case client.send <- ackPayload:
	default:
	}
}

func (h *Handler) handleMessageRead(client *Client, msg IncomingMessage) {
	var data chatDto.MessageReadData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		h.sendError(client, msg.RequestID, "INVALID_PAYLOAD", "invalid read data")
		return
	}

	isParticipant, err := h.chatSvc.IsParticipant(client.UserID(), data.ConversationID)
	if err != nil || !isParticipant {
		return
	}

	err = h.chatSvc.MarkAsRead(client.UserID(), data.ConversationID, data.MessageID)
	if err != nil {
		return
	}

	h.hub.BroadcastToRoomExcept(data.ConversationID, chatDto.EVENT_MESSAGE_READ, chatDto.MessageReadData{
		ConversationID: data.ConversationID,
		MessageID:      data.MessageID,
	}, client.UserID())
}

func (h *Handler) handleTyping(client *Client, msg IncomingMessage, isTyping bool) {
	var data chatDto.TypingData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return
	}

	isParticipant, err := h.chatSvc.IsParticipant(client.UserID(), data.ConversationID)
	if err != nil || !isParticipant {
		return
	}

	eventType := chatDto.EVENT_TYPING_STOP
	if isTyping {
		eventType = chatDto.EVENT_TYPING_START
	}

	h.hub.BroadcastToRoomExcept(data.ConversationID, eventType, chatDto.TypingData{
		ConversationID: data.ConversationID,
		UserID:         client.UserID(),
	}, client.UserID())
}

func (h *Handler) handleConversationJoin(client *Client, msg IncomingMessage) {
	var data chatDto.ConversationJoinData
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		h.sendError(client, msg.RequestID, "INVALID_PAYLOAD", "invalid join data")
		return
	}

	if data.ConversationID == "" {
		h.sendError(client, msg.RequestID, "INVALID_PAYLOAD", "conversation_id required")
		return
	}

	isParticipant, err := h.chatSvc.IsParticipant(client.UserID(), data.ConversationID)
	if err != nil {
		h.sendError(client, msg.RequestID, "SERVER_ERROR", "failed to verify participant")
		return
	}
	if !isParticipant {
		h.sendError(client, msg.RequestID, "FORBIDDEN", "you are not a participant of this conversation")
		return
	}

	h.hub.JoinClientToRoom(client, data.ConversationID)
}

func (h *Handler) sendError(client *Client, requestID string, code string, message string) {
	payload, _ := json.Marshal(map[string]interface{}{
		"type": chatDto.EVENT_ERROR,
		"data": chatDto.WSErrorData{
			RequestID: requestID,
			Code:      code,
			Message:   message,
		},
	})
	select {
	case client.send <- payload:
	default:
	}
}

func (h *Handler) JoinUserToRoom(conversationID string, userID string) bool {
	isParticipant, err := h.chatSvc.IsParticipant(userID, conversationID)
	if err != nil || !isParticipant {
		return false
	}

	h.hub.JoinUserToRoom(userID, conversationID)
	return true
}

func (h *Handler) handlePresenceChange(userID string, isOnline bool, lastSeenAt *time.Time, roomIDs []string) {
	ctx := context.Background()

	if !isOnline && lastSeenAt != nil {
		err := h.userRepo.UpdateLastSeenAt(ctx, h.db, userID, lastSeenAt)
		if err != nil {
			log.Printf("[WS] failed to update last_seen_at: %v", err)
		}
	}

	h.hub.BroadcastToRooms(roomIDs, chatDto.EVENT_PRESENCE_CHANGED, chatDto.PresenceChangedData{
		UserID:     userID,
		IsOnline:   isOnline,
		LastSeenAt: lastSeenAt,
	}, userID)
}
