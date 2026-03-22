package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Payment struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrderID         uuid.UUID       `gorm:"type:uuid;not null;index"`
	MidtransOrderID string          `gorm:"type:varchar(100);not null;uniqueIndex"`
	Amount          decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	PaymentMethod   string          `gorm:"type:varchar(50)"`
	PaymentType     string          `gorm:"type:varchar(50)"`
	Status          string          `gorm:"type:varchar(50);not null;default:'Pending'"`
	MidtransTransID string          `gorm:"type:varchar(100)"`
	SnapToken       string          `gorm:"type:varchar(255)"`
	PaidAt          *time.Time      `gorm:"type:timestamp with time zone"`
	ExpiredAt       *time.Time      `gorm:"type:timestamp with time zone"`

	Order Order `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	Timestamp
}
