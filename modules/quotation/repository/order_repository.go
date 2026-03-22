package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

type (
	OrderRepository interface {
		Create(ctx context.Context, tx *gorm.DB, order entities.Order) (entities.Order, error)
	}

	orderRepository struct {
		db *gorm.DB
	}
)

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, tx *gorm.DB, order entities.Order) (entities.Order, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&order).Error; err != nil {
		return entities.Order{}, err
	}

	return order, nil
}
