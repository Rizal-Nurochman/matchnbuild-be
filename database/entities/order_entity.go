package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Order struct {
	ID            uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	QuotationID   uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex"`
	ClientID      uuid.UUID       `gorm:"type:uuid;not null;index"`
	DesignerID    uuid.UUID       `gorm:"type:uuid;not null;index"`
	TotalAmount   decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	PaymentStatus string          `gorm:"type:varchar(50);not null;default:'Unpaid'"`
	WorkStatus    string          `gorm:"type:varchar(50);not null;default:'Active'"`
	CreatedAt     time.Time       `gorm:"type:timestamp with time zone;autoCreateTime"`
	CompletedAt   *time.Time      `gorm:"type:timestamp with time zone"`

	Quotation    Quotation       `gorm:"foreignKey:QuotationID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Client       User            `gorm:"foreignKey:ClientID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Designer     DesignerProfile `gorm:"foreignKey:DesignerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Deliverables []Deliverable   `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Review       *Review         `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
