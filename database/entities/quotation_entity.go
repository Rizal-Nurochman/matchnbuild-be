package entities

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Quotation struct {
	ID               uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ProjectRequestID uuid.UUID       `gorm:"type:uuid;not null;index"`
	DesignerID       uuid.UUID       `gorm:"type:uuid;not null;index"`
	ScopeOfWork      string          `gorm:"type:text;not null"`
	OfferedPrice     decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	DurationDays     int             `gorm:"type:int;not null"`
	Status           string          `gorm:"type:varchar(50);not null;default:'Pending'"`

	ProjectRequest ProjectRequest  `gorm:"foreignKey:ProjectRequestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Designer       DesignerProfile `gorm:"foreignKey:DesignerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Order          *Order          `gorm:"foreignKey:QuotationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Timestamp
}
