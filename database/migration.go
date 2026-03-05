package database

import (
	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&entities.Migration{},
		&entities.User{},
		&entities.RefreshToken{},
		&entities.DesignerProfile{},
		&entities.ProjectType{},
		&entities.DesignItem{},
		&entities.ProjectRequest{},
		&entities.Quotation{},
		&entities.Order{},
		&entities.Deliverable{},
		&entities.Review{},
		&entities.Conversation{},
		&entities.Message{},
	); err != nil {
		return err
	}

	manager := NewMigrationManager(db)
	return manager.Run()
}
