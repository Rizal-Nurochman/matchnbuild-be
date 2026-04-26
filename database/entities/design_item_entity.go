package entities

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DesignItem struct {
	ID             uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	DesignerID     uuid.UUID       `gorm:"type:uuid;not null;index"`
	Title          string          `gorm:"type:varchar(200);not null"`
	Description    string          `gorm:"type:text"`
	Style					 string					 `gorm:"type:varchar(50);not null"`
	LandArea			 float64				 `gorm:"type:float;not null"`
	BuildingArea	 float64				 `gorm:"type:float;not null"`
	NumFloors			 int						 `gorm:"type:int;not null"`
	NumBedrooms		 *int						 `gorm:"type:int"`
	EstimatedBudget decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	PriceStartFrom decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0"`
	ImageURL       string          `gorm:"type:varchar(255)"`

	Designer    DesignerProfile `gorm:"foreignKey:DesignerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Features 		[]DesignItemFeature

	Timestamp
}
