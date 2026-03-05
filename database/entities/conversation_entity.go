package entities

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ProjectRequestID uuid.UUID  `gorm:"type:uuid;not null;index"`
	OrderID          *uuid.UUID `gorm:"type:uuid;index"`
	CreatedAt        time.Time  `gorm:"type:timestamp with time zone;autoCreateTime"`

	ProjectRequest ProjectRequest `gorm:"foreignKey:ProjectRequestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Order          *Order         `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Messages       []Message      `gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
