package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

type (
	QuotationRepository interface {
		Create(ctx context.Context, tx *gorm.DB, quotation entities.Quotation) (entities.Quotation, error)
		GetByID(ctx context.Context, tx *gorm.DB, id string) (entities.Quotation, error)
		GetByProjectRequestID(ctx context.Context, tx *gorm.DB, projectRequestID string) (entities.Quotation, error)
		Update(ctx context.Context, tx *gorm.DB, quotation entities.Quotation) (entities.Quotation, error)
	}

	quotationRepository struct {
		db *gorm.DB
	}
)

func NewQuotationRepository(db *gorm.DB) QuotationRepository {
	return &quotationRepository{db: db}
}

func (r *quotationRepository) Create(ctx context.Context, tx *gorm.DB, quotation entities.Quotation) (entities.Quotation, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&quotation).Error; err != nil {
		return entities.Quotation{}, err
	}

	return quotation, nil
}

func (r *quotationRepository) GetByID(ctx context.Context, tx *gorm.DB, id string) (entities.Quotation, error) {
	if tx == nil {
		tx = r.db
	}

	var quotation entities.Quotation
	if err := tx.WithContext(ctx).Preload("ProjectRequest").Where("id = ?", id).Take(&quotation).Error; err != nil {
		return entities.Quotation{}, err
	}

	return quotation, nil
}

func (r *quotationRepository) GetByProjectRequestID(ctx context.Context, tx *gorm.DB, projectRequestID string) (entities.Quotation, error) {
	if tx == nil {
		tx = r.db
	}

	var quotation entities.Quotation
	if err := tx.WithContext(ctx).Where("project_request_id = ?", projectRequestID).Take(&quotation).Error; err != nil {
		return entities.Quotation{}, err
	}

	return quotation, nil
}

func (r *quotationRepository) Update(ctx context.Context, tx *gorm.DB, quotation entities.Quotation) (entities.Quotation, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Updates(&quotation).Error; err != nil {
		return entities.Quotation{}, err
	}

	return quotation, nil
}
