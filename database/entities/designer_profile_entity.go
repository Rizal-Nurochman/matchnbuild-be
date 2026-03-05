package entities

import (
	"github.com/google/uuid"
)

type DesignerProfile struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID            uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Bio               string    `gorm:"type:text"`
	ExperienceYears   int       `gorm:"type:int;default:0"`
	IsVerified        bool      `gorm:"default:false"`
	IsAvailable       bool      `gorm:"default:true"`
	Location          string    `gorm:"type:varchar(255)"`
	BankAccountNumber string    `gorm:"type:varchar(50)"`

	User            User             `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	DesignItems     []DesignItem     `gorm:"foreignKey:DesignerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ProjectRequests []ProjectRequest `gorm:"foreignKey:DesignerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Timestamp
}
