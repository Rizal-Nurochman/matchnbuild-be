package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name               string     `gorm:"type:varchar(100);not null"`
	Email              string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password           string     `gorm:"type:varchar(255);not null"`
	Role               string     `gorm:"type:varchar(50);not null;default:'client'"`
	ProfilePicture     string     `gorm:"type:varchar(255)"`
	IsVerified         bool       `gorm:"default:false;not null"`
	VerificationCode   string     `gorm:"type:varchar(6)"`
	VerificationExpiry *time.Time `gorm:"type:timestamp with time zone"`

	DesignerProfile *DesignerProfile `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ProjectRequests []ProjectRequest `gorm:"foreignKey:ClientID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Messages        []Message        `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Timestamp
}
