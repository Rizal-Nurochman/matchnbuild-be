package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRepository interface {
	GetOrderByID(ctx context.Context, tx *gorm.DB, orderID string, lock bool) (entities.Order, error)
	UpdateOrder(ctx context.Context, tx *gorm.DB, order entities.Order) error
	Create(ctx context.Context, tx *gorm.DB, payment entities.Payment) (entities.Payment, error)
	Update(ctx context.Context, tx *gorm.DB, payment entities.Payment) error
	GetLatestByOrderID(ctx context.Context, tx *gorm.DB, orderID string) (entities.Payment, error)
	GetByMidtransOrderID(ctx context.Context, tx *gorm.DB, midtransOrderID string, lock bool) (entities.Payment, error)
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) GetOrderByID(ctx context.Context, tx *gorm.DB, orderID string, lock bool) (entities.Order, error) {
	if tx == nil {
		tx = r.db
	}

	query := tx.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	var order entities.Order
	err := query.Preload("Client").Where("id = ?", orderID).Take(&order).Error
	return order, err
}

func (r *paymentRepository) UpdateOrder(ctx context.Context, tx *gorm.DB, order entities.Order) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Model(&entities.Order{}).
		Where("id = ?", order.ID).
		Update("payment_status", order.PaymentStatus).Error
}

func (r *paymentRepository) Create(ctx context.Context, tx *gorm.DB, payment entities.Payment) (entities.Payment, error) {
	if tx == nil {
		tx = r.db
	}
	err := tx.WithContext(ctx).Create(&payment).Error
	return payment, err
}

func (r *paymentRepository) Update(ctx context.Context, tx *gorm.DB, payment entities.Payment) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Save(&payment).Error
}

func (r *paymentRepository) GetLatestByOrderID(ctx context.Context, tx *gorm.DB, orderID string) (entities.Payment, error) {
	if tx == nil {
		tx = r.db
	}

	var payment entities.Payment
	err := tx.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Take(&payment).Error
	return payment, err
}

func (r *paymentRepository) GetByMidtransOrderID(ctx context.Context, tx *gorm.DB, midtransOrderID string, lock bool) (entities.Payment, error) {
	if tx == nil {
		tx = r.db
	}

	query := tx.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	var payment entities.Payment
	err := query.Where("midtrans_order_id = ?", midtransOrderID).Take(&payment).Error
	return payment, err
}
