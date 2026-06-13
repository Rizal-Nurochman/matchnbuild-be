package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

type (
	MessageRepository interface {
		Create(ctx context.Context, tx *gorm.DB, message entities.Message) (entities.Message, error)
		GetByConversationID(ctx context.Context, conversationID string, limit int) ([]entities.Message, error)
	}

	messageRepository struct {
		db *gorm.DB
	}
)

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func(r *messageRepository) Create(ctx context.Context, tx *gorm.DB, message entities.Message) (entities.Message, error) {
	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Create(&message).Error
	if err != nil {
		return entities.Message{}, err
	}

	return message, nil
}

func (r *messageRepository) GetByConversationID(ctx context.Context, conversationID string, limit int) ([]entities.Message, error) {
	var messages []entities.Message

	if err := r.db.WithContext(ctx).
		Preload("Sender").
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}