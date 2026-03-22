package entities

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProjectRequest struct {
	ID               uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ClientID         uuid.UUID       `gorm:"type:uuid;not null;index"`
	DesignerID       uuid.UUID       `gorm:"type:uuid;not null;index"`
	Description      string          `gorm:"type:text;not null"`
	InitialBudget    decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0"`
	AreaSize         float64         `gorm:"type:float;not null;default:0"`
	LocationPhotoURL string          `gorm:"type:varchar(255)"`
	LayoutSketchURL  string          `gorm:"type:varchar(255)"`
	Status           string          `gorm:"type:varchar(50);not null;default:'Open'"`

	Client        User            `gorm:"foreignKey:ClientID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Designer      DesignerProfile `gorm:"foreignKey:DesignerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Quotations    []Quotation     `gorm:"foreignKey:ProjectRequestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Conversations []Conversation  `gorm:"foreignKey:ProjectRequestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Timestamp
}
