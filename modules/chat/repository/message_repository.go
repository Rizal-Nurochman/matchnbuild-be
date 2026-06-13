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
		GetByConversationID(ctx context.Context, tx *gorm.DB, conversationID string, beforeID string, afterID string, limit int) ([]entities.Message, error)
		GetUnreadCount(ctx context.Context, tx *gorm.DB, conversationID string, userID string) (int64, error)
		GetTotalUnreadCount(ctx context.Context, tx *gorm.DB, userID string) (int64, error)
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

func (r *messageRepository) GetByConversationID(ctx context.Context, tx *gorm.DB, conversationID string, beforeID string, afterID string, limit int) ([]entities.Message, error) {
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

	if afterID != "" {
		var afterMessage entities.Message
		if err := tx.WithContext(ctx).Where("id = ?", afterID).Take(&afterMessage).Error; err == nil {
			query = query.Where("created_at > ?", afterMessage.CreatedAt)
		}
	}

	if err := query.Order("created_at ASC").Limit(limit).Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *messageRepository) GetUnreadCount(ctx context.Context, tx *gorm.DB, conversationID string, userID string) (int64, error) {
	if tx == nil {
		tx = r.db
	}

	var lastReadMessageID *string
	err := tx.WithContext(ctx).
		Table("conversation_participants").
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Select("last_read_message_id").
		Scan(&lastReadMessageID).Error
	if err != nil {
		return 0, err
	}

	query := tx.WithContext(ctx).
		Model(&entities.Message{}).
		Where("conversation_id = ? AND sender_id != ?", conversationID, userID)

	if lastReadMessageID != nil && *lastReadMessageID != "" {
		query = query.Where("created_at > (SELECT created_at FROM messages WHERE id = ?)", *lastReadMessageID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *messageRepository) GetTotalUnreadCount(ctx context.Context, tx *gorm.DB, userID string) (int64, error) {
	if tx == nil {
		tx = r.db
	}

	var total int64
	err := tx.WithContext(ctx).
		Raw(`
			SELECT COALESCE(SUM(sub.unread_count), 0)
			FROM (
				SELECT COUNT(m.id) as unread_count
				FROM messages m
				JOIN conversation_participants cp ON cp.conversation_id = m.conversation_id
				WHERE cp.user_id = ? AND m.sender_id != ?
				AND (cp.last_read_message_id IS NULL OR m.created_at > (
					SELECT created_at FROM messages WHERE id = cp.last_read_message_id
				))
				GROUP BY m.conversation_id
			) sub
		`, userID, userID).
		Scan(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}
