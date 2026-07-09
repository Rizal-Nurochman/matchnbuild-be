package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type (
	designItemRepository struct {
		db *gorm.DB
	}

	DesignItemRepository interface {
		Create(ctx context.Context, tx *gorm.DB, designItemReq entities.DesignItem) (entities.DesignItem, error)
		GetAll(ctx context.Context, tx *gorm.DB) ([]entities.DesignItem, error)
		GetAllFeatures(ctx context.Context, tx *gorm.DB) ([]entities.Feature, error)
		GetByID(ctx context.Context, tx *gorm.DB, designItemID string) (entities.DesignItem, error)
		GetByCategory(ctx context.Context, tx *gorm.DB, category string) ([]entities.Feature, error)
		GetByDesignerID(ctx context.Context, tx *gorm.DB, designerID string) ([]entities.DesignItem, error)
		GetRecommended(ctx context.Context, tx *gorm.DB, style string, budgetMin decimal.Decimal, budgetMax decimal.Decimal, location string, limit int) ([]entities.DesignItem, error)
		Update(ctx context.Context, tx *gorm.DB, designItemReq entities.DesignItem) (entities.DesignItem, error)
		Delete(ctx context.Context, tx *gorm.DB, designItemID string) error
	}
)

func NewDesignItemRepository(db *gorm.DB) DesignItemRepository {
	return &designItemRepository{db: db}
}

func (r *designItemRepository) Create(ctx context.Context, tx *gorm.DB, designItemReq entities.DesignItem) (entities.DesignItem, error) {
	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Create(&designItemReq).Error
	if err != nil {
		return entities.DesignItem{}, err
	}

	return designItemReq, nil
}

func (r *designItemRepository) GetAll(ctx context.Context, tx *gorm.DB) ([]entities.DesignItem, error) {
	if tx == nil {
		tx = r.db
	}

	var designItems []entities.DesignItem
	err := tx.WithContext(ctx).Preload("Designer").Preload("Designer.User").Preload("Features.Feature").Find(&designItems).Error
	if err != nil {
		return nil, err
	}

	return designItems, nil
}

func (r *designItemRepository) GetAllFeatures(ctx context.Context, tx *gorm.DB) ([]entities.Feature, error) {
	if tx == nil {
		tx = r.db
	}

	var features []entities.Feature
	err := tx.WithContext(ctx).Find(&features).Error
	if err != nil {
		return nil, err
	}

	return features, nil
}

func (r *designItemRepository) GetByCategory(ctx context.Context, tx *gorm.DB, category string) ([]entities.Feature, error) {
	if tx == nil {
		tx = r.db
	}

	var features []entities.Feature
	err := tx.WithContext(ctx).Where("category = ?", category).Take(&features).Error
	if err != nil {
		return nil, err
	}

	return features, nil
}

func (r *designItemRepository) GetByID(ctx context.Context, tx *gorm.DB, designItemID string) (entities.DesignItem, error) {
	if tx == nil {
		tx = r.db
	}

	var designItem entities.DesignItem
	err := tx.WithContext(ctx).Preload("Designer").Preload("Designer.User").Preload("Features.Feature").Where("id = ?", designItemID).Take(&designItem).Error
	if err != nil {
		return designItem, err
	}

	return designItem, nil
}

func (r *designItemRepository) GetByDesignerID(ctx context.Context, tx *gorm.DB, designerID string) ([]entities.DesignItem, error) {
	if tx == nil {
		tx = r.db
	}

	var designItems []entities.DesignItem
	err := tx.WithContext(ctx).Preload("Designer").Preload("Designer.User").Preload("Features.Feature").Where("designer_id = ?", designerID).Find(&designItems).Error
	if err != nil {
		return designItems, err
	}

	return designItems, nil
}

func (r *designItemRepository) Update(ctx context.Context, tx *gorm.DB, designItemReq entities.DesignItem) (entities.DesignItem, error) {
	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Updates(&designItemReq).Error
	if err != nil {
		return entities.DesignItem{}, err
	}

	return designItemReq, nil
}

func (r *designItemRepository) Delete(ctx context.Context, tx *gorm.DB, designItemID string) error {
	if tx == nil {
		tx = r.db
	}

	// design_item_features FK is RESTRICT, so remove join rows before the parent.
	return tx.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("design_item_id = ?", designItemID).
			Delete(&entities.DesignItemFeature{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", designItemID).Delete(&entities.DesignItem{}).Error
	})
}

func (r *designItemRepository) GetRecommended(ctx context.Context, tx *gorm.DB, style string, budgetMin decimal.Decimal, budgetMax decimal.Decimal, location string, limit int) ([]entities.DesignItem, error) {
	if tx == nil {
		tx = r.db
	}

	var ids []string
	err := tx.WithContext(ctx).
		Raw(`SELECT id FROM design_items
			WHERE style = ? OR estimated_budget BETWEEN ? AND ? OR location = ?
			ORDER BY
				(CASE WHEN style = ? THEN 2 ELSE 0 END) +
				(CASE WHEN estimated_budget BETWEEN ? AND ? THEN 1 ELSE 0 END) +
				(CASE WHEN location = ? THEN 1 ELSE 0 END) DESC,
				created_at DESC
			LIMIT ?`,
			style, budgetMin, budgetMax, location,
			style, budgetMin, budgetMax, location,
			limit,
		).
		Scan(&ids).Error
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		err = tx.WithContext(ctx).
			Raw(`SELECT id FROM design_items ORDER BY created_at DESC LIMIT ?`, limit).
			Scan(&ids).Error
		if err != nil {
			return nil, err
		}
	}

	if len(ids) == 0 {
		return []entities.DesignItem{}, nil
	}

	var items []entities.DesignItem
	err = tx.WithContext(ctx).
		Preload("Designer").Preload("Designer.User").Preload("Features.Feature").
		Where("id IN ?", ids).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	orderMap := make(map[string]int, len(ids))
	for i, id := range ids {
		orderMap[id] = i
	}

	ordered := make([]entities.DesignItem, len(items))
	for _, item := range items {
		idx := orderMap[item.ID.String()]
		ordered[idx] = item
	}

	return ordered, nil
}
