package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	ConversationParticipantRepository interface {
		CreateMany(ctx context.Context, tx *gorm.DB, participants []entities.ConversationParticipant) error
		IsParticipant(ctx context.Context, tx *gorm.DB, conversationID string, userID string) (bool, error)
		GetByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]entities.ConversationParticipant, error)
		UpdateLastReadMessage(ctx context.Context, tx *gorm.DB, conversationID string, userID string, messageID string) error
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

func (r *conversationParticipantRepository) IsParticipant(ctx context.Context, tx *gorm.DB, conversationID string, userID string) (bool, error) {
	if tx == nil {
		tx = r.db
	}
	var count int64

	if err := tx.WithContext(ctx).
		Model(&entities.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Count(&count).
		Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *conversationParticipantRepository) GetByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]entities.ConversationParticipant, error) {
	if tx == nil {
		tx = r.db
	}

	var participants []entities.ConversationParticipant
	if err := tx.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&participants).Error; err != nil {
		return nil, err
	}

	return participants, nil
}

func (r *conversationParticipantRepository) UpdateLastReadMessage(ctx context.Context, tx *gorm.DB, conversationID string, userID string, messageID string) error {
	if tx == nil {
		tx = r.db
	}

	parsed, err := uuid.Parse(messageID)
	if err != nil {
		return err
	}

	if err := tx.WithContext(ctx).
		Model(&entities.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Where(`(
			last_read_message_id IS NULL
			OR (SELECT created_at FROM messages WHERE id = ?) >
			   (SELECT created_at FROM messages WHERE id = last_read_message_id)
		)`, parsed).
		Update("last_read_message_id", parsed).Error; err != nil {
		return err
	}

	return nil
}