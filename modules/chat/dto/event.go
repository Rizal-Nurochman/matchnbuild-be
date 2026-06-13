package dto

import (
	"encoding/json"
	"time"
)

const (
	EVENT_MESSAGE_SEND     	= "message.send"
	EVENT_MESSAGE_CREATED  	= "message.created"
	EVENT_MESSAGE_DELIVERED = "message.delivered"
	EVENT_MESSAGE_READ     	= "message.read"
	EVENT_TYPING_START     	= "typing.start"
	EVENT_TYPING_STOP      	= "typing.stop"
	EVENT_PRESENCE_CHANGED 	= "presence.changed"
	EVENT_ERROR            	= "error"
)

type (
	WSEvent struct {
		Type      string          `json:"type"`
		RequestID string          `json:"request_id,omitempty"`
		Data      json.RawMessage `json:"data"`
	}

	WSResponse struct {
		Type string `json:"type"`
		Data any    `json:"data"`
	}

	MessageSendData struct {
		ConversationID  string `json:"conversation_id"`
		ClientMessageID string `json:"client_message_id"`
		MessageType     string `json:"message_type"`
		Text            string `json:"text"`
		AttachmentURL   string `json:"attachment_url,omitempty"`
	}

	MessageCreatedData struct {
		ID             string    `json:"id"`
		ConversationID string    `json:"conversation_id"`
		SenderID       string    `json:"sender_id"`
		SenderName     string    `json:"sender_name"`
		ClientMessageID *string  `json:"client_message_id,omitempty"`
		MessageType    string    `json:"message_type"`
		Text           string    `json:"text"`
		AttachmentURL  string    `json:"attachment_url,omitempty"`
		CreatedAt      time.Time `json:"created_at"`
	}

	MessageDeliveredData struct {
		ConversationID string `json:"conversation_id"`
		MessageID      string `json:"message_id"`
	}

	MessageReadData struct {
		ConversationID string `json:"conversation_id"`
		MessageID      string `json:"message_id"`
	}

	TypingData struct {
		ConversationID string `json:"conversation_id"`
		UserID         string `json:"user_id"`
	}

	PresenceChangedData struct {
		UserID     string `json:"user_id"`
		IsOnline   bool   `json:"is_online"`
		LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	}

	WSErrorData struct {
		RequestID string `json:"request_id,omitempty"`
		Code      string `json:"code"`
		Message   string `json:"message"`
	}
)
