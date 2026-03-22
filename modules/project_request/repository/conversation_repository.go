package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

type (
	ConversationRepository interface {
		Create(ctx context.Context, tx *gorm.DB, conversation entities.Conversation) (entities.Conversation, error)
		GetByProjectRequestID(ctx context.Context, tx *gorm.DB, projectRequestID string) (entities.Conversation, error)
	}

	conversationRepository struct {
		db *gorm.DB
	}
)

func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(ctx context.Context, tx *gorm.DB, conversation entities.Conversation) (entities.Conversation, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&conversation).Error; err != nil {
		return entities.Conversation{}, err
	}

	return conversation, nil
}

func (r *conversationRepository) GetByProjectRequestID(ctx context.Context, tx *gorm.DB, projectRequestID string) (entities.Conversation, error) {
	if tx == nil {
		tx = r.db
	}

	var conversation entities.Conversation
	if err := tx.WithContext(ctx).Where("project_request_id = ?", projectRequestID).Take(&conversation).Error; err != nil {
		return entities.Conversation{}, err
	}

	return conversation, nil
}
