package entities

import (
	"github.com/google/uuid"
)

type ConversationParticipant struct {
	ConversationID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	Role           string    `gorm:"type:varchar(50);not null"`

	LastReadMessageID *uuid.UUID `gorm:"type:uuid;index"`

	Conversation Conversation `gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User         User         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	LastReadMessage *Message  `gorm:"foreignKey:LastReadMessageID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	Timestamp
}