package entities

import (
	"time"

	"github.com/google/uuid"
)

type Deliverable struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrderID        uuid.UUID `gorm:"type:uuid;not null;index"`
	FileURL        string    `gorm:"type:varchar(255);not null"`
	FileType       string    `gorm:"type:varchar(50);not null"`
	Description    string    `gorm:"type:text"`
	RevisionNumber int       `gorm:"type:int;not null;default:1"`
	CreatedAt      time.Time `gorm:"type:timestamp with time zone;autoCreateTime"`

	Order Order `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
