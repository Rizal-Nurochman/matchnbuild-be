package entities

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DesignItem struct {
	ID             uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	DesignerID     uuid.UUID       `gorm:"type:uuid;not null;index"`
	Category		   string					 `gorm:"type:varchar(20);not null"`
	Title          string          `gorm:"type:varchar(200);not null"`
	Description    string          `gorm:"type:text"`
	Style					 string					 `gorm:"type:varchar(50);not null"`
	LandAreaMin		 *float64			 `gorm:"type:float"`
	LandAreaMax		 *float64			 `gorm:"type:float"`
	BuildingArea	 *float64				 `gorm:"type:float"`
	NumFloors			 *int						 `gorm:"type:int"`
	NumBedrooms		 *int						 `gorm:"type:int"`
	RoomType			 *string					 `gorm:"type:varchar(50)"`
	RoomArea			 *float64				 `gorm:"type:float"`
	EstimatedBudget decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	PriceStartFrom decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0"`
	ImageURL       string          `gorm:"type:text"`
	Location       string          `gorm:"type:varchar(100)"`

	Designer    DesignerProfile `gorm:"foreignKey:DesignerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Features 		[]DesignItemFeature

	Timestamp
}
