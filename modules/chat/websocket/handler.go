package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	chatDto "github.com/Rizal-Nurochman/matchnbuild/modules/chat/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/chat/service"
	"github.com/Rizal-Nurochman/matchnbuild/modules/user/repository"
	authService "github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub        *Hub
	chatSvc    service.ChatService
	userRepo   repository.UserRepository
	jwtService authService.JWTService
	db         *gorm.DB
}

func NewHandler(hub *Hub, chatSvc service.ChatService, userRepo repository.UserRepository, jwtService authService.JWTService, db *gorm.DB) *Handler {
	h := &Handler{
		hub:        hub,
		chatSvc:    chatSvc,
		userRepo:   userRepo,
		jwtService: jwtService,
		db:         db,
	}

	hub.SetOnPresence(h.handlePresenceChange)
	return h
}

func (h *Handler) HandleWebSocket(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	userID, err := h.jwtService.GetUserIDByToken(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("[WS] upgrade failed: %v", err)
		return
	}

	client := NewClient(h.hub, conn, userID)
	h.hub.register <- client

	go client.WritePump()
	go client.ReadPump(h.handleEvent)
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
		"type": chatDto.EVENT_MESSAGE_CREATED,
		"data": createdData,
		"request_id": msg.RequestID,
	})
	select {
	case client.send <- ackPayload:
	default:
	}

	h.hub.BroadcastToRoomExcept(data.ConversationID, chatDto.EVENT_MESSAGE_CREATED, createdData, client.UserID())
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

func (h *Handler) handlePresenceChange(userID string, isOnline bool, lastSeenAt *time.Time) {
	ctx := context.Background()

	if !isOnline && lastSeenAt != nil {
		err := h.userRepo.UpdateLastSeenAt(ctx, h.db, userID, lastSeenAt)
		if err != nil {
			log.Printf("[WS] failed to update last_seen_at: %v", err)
		}
	}

	h.hub.BroadcastToAllUsers(chatDto.EVENT_PRESENCE_CHANGED, chatDto.PresenceChangedData{
		UserID:     userID,
		IsOnline:   isOnline,
		LastSeenAt: lastSeenAt,
	})
}
