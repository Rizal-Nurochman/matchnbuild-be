package dto

import (
	"errors"
	"time"
)

const (
	MESSAGE_FAILED_GET_CONVERSATIONS    = "failed get conversations"
	MESSAGE_SUCCESS_GET_CONVERSATIONS    = "success get conversations"
	MESSAGE_FAILED_GET_MESSAGES         = "failed get messages"
	MESSAGE_SUCCESS_GET_MESSAGES         = "success get messages"
	MESSAGE_FAILED_SEND_MESSAGE         = "failed send message"
	MESSAGE_SUCCESS_SEND_MESSAGE         = "success send message"
	MESSAGE_FAILED_GET_UNREAD_COUNT     = "failed get unread count"
	MESSAGE_SUCCESS_GET_UNREAD_COUNT     = "success get unread count"
)

var (
	ErrConversationNotFound       = errors.New("conversation not found")
	ErrNotConversationParticipant = errors.New("you are not a participant of this conversation")
	ErrMessageNotFound            = errors.New("message not found")
	ErrInvalidMessageType         = errors.New("invalid message type")
	ErrDuplicateMessage           = errors.New("message already sent")
	ErrEmptyMessage               = errors.New("message text or attachment is required")
	ErrMessageTextTooLong         = errors.New("message text exceeds 5000 characters")
	ErrAttachmentURLTooLong       = errors.New("attachment url exceeds 255 characters")
	ErrClientMessageIDTooLong     = errors.New("client_message_id exceeds 36 characters")
)

const (
	MaxMessageTextLength   = 5000
	MaxAttachmentURLLength = 255
	MaxClientMessageIDLen  = 36
)

type (
	ConversationUserInfo struct {
		ID             string  `json:"id"`
		Name           string  `json:"name"`
		ProfilePicture *string `json:"profile_picture"`
	}

	ConversationProjectRequestInfo struct {
		ID            string  `json:"id"`
		Description   string  `json:"description"`
		Status        string  `json:"status"`
		InitialBudget float64 `json:"initial_budget"`
	}

	ConversationResponse struct {
		ID             string                          `json:"id"`
		ProjectRequestID string                        `json:"project_request_id"`
		OrderID        *string                         `json:"order_id,omitempty"`
		ProjectRequest ConversationProjectRequestInfo  `json:"project_request"`
		Client         ConversationUserInfo            `json:"client"`
		Designer       ConversationUserInfo            `json:"designer"`
		CreatedAt      time.Time                       `json:"created_at"`
	}

	SendMessageRequest struct {
		ClientMessageID string `json:"client_message_id" binding:"required"`
		MessageText     string `json:"message_text" binding:"required,max=5000"`
		AttachmentURL   string `json:"attachment_url" binding:"omitempty,max=255"`
		MessageType     string `json:"message_type" binding:"required,oneof=Text Image File"`
	}

	MessageResponse struct {
		ID              string    `json:"id"`
		ConversationID  string    `json:"conversation_id"`
		SenderID        string    `json:"sender_id"`
		SenderName      string    `json:"sender_name"`
		ClientMessageID *string   `json:"client_message_id,omitempty"`
		MessageText     string    `json:"message_text"`
		AttachmentURL   string    `json:"attachment_url"`
		MessageType     string    `json:"message_type"`
		CreatedAt       time.Time `json:"created_at"`
	}

	GetMessagesResponse struct {
		Messages []MessageResponse `json:"messages"`
		HasMore  bool              `json:"has_more"`
	}

	UnreadCountResponse struct {
		ConversationID string `json:"conversation_id"`
		UnreadCount    int    `json:"unread_count"`
	}

	TotalUnreadCountResponse struct {
		TotalUnreadCount int `json:"total_unread_count"`
	}
)
