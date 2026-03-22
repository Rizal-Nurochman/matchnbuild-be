package entities

import (
	"time"
)

type Migration struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Batch     int       `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone"`
}
