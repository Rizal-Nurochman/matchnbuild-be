package entities

import "github.com/google/uuid"

type DesignItemFeature struct {
	FeatureID uuid.UUID	`gorm:"type:uuid;primary_key"`
	DesignItemID uuid.UUID `gorm:"type:uuid;primary_key"`

	//Relations
	Feature    Feature    `gorm:"foreignKey:FeatureID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	DesignItem DesignItem `gorm:"foreignKey:DesignItemID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Timestamp
}