package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
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
		Update(ctx context.Context, tx *gorm.DB, designItemReq entities.DesignItem) (entities.DesignItem, error)
		Delete(ctx context.Context, tx *gorm.DB, designItemID string) (error)
	}
)

func NewDesignItemRepository(db *gorm.DB) DesignItemRepository  {
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

func (r *designItemRepository) GetAllFeatures(ctx context.Context, tx *gorm.DB) ([]entities.Feature, error)  {
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

func (r *designItemRepository) GetByCategory(ctx context.Context, tx *gorm.DB, category string) ([]entities.Feature, error)  {
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

func (r *designItemRepository) Update(ctx context.Context, tx *gorm.DB, designItemReq entities.DesignItem) (entities.DesignItem, error)  {
	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Updates(&designItemReq).Error
	if err != nil {
		return entities.DesignItem{}, err
	}

	return designItemReq, nil
}

func (r *designItemRepository) Delete(ctx context.Context, tx *gorm.DB, designItemID string) (error) {
	if tx == nil {
		tx = r.db
	}

	var designItem entities.DesignItem
	err := tx.WithContext(ctx).Where("id = ?", designItemID).Delete(&designItem).Error
	if err != nil {
		return err
	}

	return nil
}