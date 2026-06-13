package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

type (
	MessageRepository interface {
		Create(ctx context.Context, tx *gorm.DB, message entities.Message) (entities.Message, error)
		GetByID(ctx context.Context, tx *gorm.DB, id string) (entities.Message, error)
		GetByConversationID(ctx context.Context, tx *gorm.DB, conversationID string, beforeID string, limit int) ([]entities.Message, error)
	}

	messageRepository struct {
		db *gorm.DB
	}
)

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, tx *gorm.DB, message entities.Message) (entities.Message, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&message).Error; err != nil {
		return entities.Message{}, err
	}

	return message, nil
}

func (r *messageRepository) GetByID(ctx context.Context, tx *gorm.DB, id string) (entities.Message, error) {
	if tx == nil {
		tx = r.db
	}

	var message entities.Message
	if err := tx.WithContext(ctx).
		Preload("Sender").
		Where("id = ?", id).
		Take(&message).Error; err != nil {
		return entities.Message{}, err
	}

	return message, nil
}

func (r *messageRepository) GetByConversationID(ctx context.Context, tx *gorm.DB, conversationID string, beforeID string, limit int) ([]entities.Message, error) {
	if tx == nil {
		tx = r.db
	}

	var messages []entities.Message
	query := tx.WithContext(ctx).
		Preload("Sender").
		Where("conversation_id = ?", conversationID)

	if beforeID != "" {
		var beforeMessage entities.Message
		if err := tx.WithContext(ctx).Where("id = ?", beforeID).Take(&beforeMessage).Error; err == nil {
			query = query.Where("created_at < ?", beforeMessage.CreatedAt)
		}
	}

	if err := query.Order("created_at DESC").Limit(limit).Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}
