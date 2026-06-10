package database

import (
	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&entities.Migration{},
		&entities.User{},
		&entities.UserPreference{},
		&entities.RefreshToken{},
		&entities.DesignerProfile{},
		&entities.DesignItem{},
		&entities.ProjectRequest{},
		&entities.Quotation{},
		&entities.Order{},
		&entities.Payment{},
		&entities.Deliverable{},
		&entities.Review{},
		&entities.Conversation{},
		&entities.Message{},
		&entities.Feature{},
		&entities.DesignItemFeature{},
		&entities.ConversationParticipant{},
	); err != nil {
		return err
	}

	manager := NewMigrationManager(db)
	return manager.Run()
}
