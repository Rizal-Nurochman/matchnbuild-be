package entities

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UserPreference struct {
	ID 								uuid.UUID						`gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID						uuid.UUID						`gorm:"type:uuid;not null;uniqueIndex"`
	PreferredStyle		string							`gorm:"type:text"`
	BudgetMin					decimal.Decimal			`gorm:"type:decimal(15,2)"`
	BudgetMax					decimal.Decimal			`gorm:"type:decimal(15,2)"`
	PreferredLocation string							`gorm:"type:varchar(100)"`
	IsOnboarded				bool								`gorm:"default:false"`

	//Relations
	User						User									`gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Timestamp
}