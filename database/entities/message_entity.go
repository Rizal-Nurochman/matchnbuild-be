package entities

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;index"`
	SenderID       uuid.UUID `gorm:"type:uuid;not null;index"`
	MessageText    string    `gorm:"type:text"`
	AttachmentURL  string    `gorm:"type:varchar(255)"`
	MessageType    string    `gorm:"type:varchar(50);not null;default:'Text'"`
	CreatedAt      time.Time `gorm:"type:timestamp with time zone;autoCreateTime"`

	Conversation Conversation `gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Sender       User         `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
