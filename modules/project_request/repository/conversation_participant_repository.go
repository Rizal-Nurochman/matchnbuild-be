package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

type (
	ConversationParticipantRepository interface {
	CreateMany(ctx context.Context, tx *gorm.DB, participants []entities.ConversationParticipant) error
}

	conversationParticipantRepository struct {
		db *gorm.DB
	}
)

func NewConversationParticipantRepository(db *gorm.DB) ConversationParticipantRepository {
	return &conversationParticipantRepository{db: db}
}

func (r *conversationParticipantRepository) CreateMany(ctx context.Context, tx *gorm.DB, participants []entities.ConversationParticipant) error {
	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Create(&participants).Error
	if err != nil {
		return err
	}

	return nil
}