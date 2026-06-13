package entities

import (
	"github.com/google/uuid"
)

type Message struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ConversationID  uuid.UUID `gorm:"type:uuid;not null;index"`
	SenderID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_sender_client_msg"`
	ClientMessageID *string   `gorm:"type:varchar(36);uniqueIndex:idx_sender_client_msg"`
	MessageText     string    `gorm:"type:text"`
	AttachmentURL   string    `gorm:"type:varchar(255)"`
	MessageType     string    `gorm:"type:varchar(50);not null;default:'Text'"`
	Timestamp

	Conversation Conversation `gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Sender       User         `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
