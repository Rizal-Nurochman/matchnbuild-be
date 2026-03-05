package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DesignItem struct {
	ID             uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	DesignerID     uuid.UUID       `gorm:"type:uuid;not null;index"`
	Title          string          `gorm:"type:varchar(200);not null"`
	Description    string          `gorm:"type:text"`
	ProjectTypeID  uuid.UUID       `gorm:"type:uuid;not null;index"`
	PriceStartFrom decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0"`
	ImageURL       string          `gorm:"type:varchar(255)"`
	CreatedAt      time.Time       `gorm:"type:timestamp with time zone;autoCreateTime"`

	Designer    DesignerProfile `gorm:"foreignKey:DesignerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ProjectType ProjectType     `gorm:"foreignKey:ProjectTypeID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}
