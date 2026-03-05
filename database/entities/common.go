package entities

import (
	"time"
)

type Timestamp struct {
	CreatedAt time.Time `gorm:"type:timestamp with time zone;autoCreateTime"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone;autoUpdateTime"`
}
