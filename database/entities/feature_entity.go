package entities

import "github.com/google/uuid"

type Feature struct {
	ID 							uuid.UUID			`gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name						string				`gorm:"type:varchar(50);not null"`
	Category 				string    `gorm:"type:varchar(20);not null"`

	//Relations
	DesignItems	[]DesignItemFeature
	Timestamp
}