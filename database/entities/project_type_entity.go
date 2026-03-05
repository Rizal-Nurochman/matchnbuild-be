package entities

import (
	"github.com/google/uuid"
)

type ProjectType struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name string    `gorm:"type:varchar(100);not null;uniqueIndex"`

	DesignItems []DesignItem `gorm:"foreignKey:ProjectTypeID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}
