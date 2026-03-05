package entities

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Rating    int       `gorm:"type:int;not null;check:rating >= 1 AND rating <= 5"`
	Comment   string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone;autoCreateTime"`

	Order Order `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
