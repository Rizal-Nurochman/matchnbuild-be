package repository

import (
    "context"
    "github.com/Rizal-Nurochman/matchnbuild/database/entities"
    "gorm.io/gorm"
)

type(
	DesignerInterface interface {
		GetByID(ctx context.Context, tx *gorm.DB, id string) (entities.DesignerProfile, error)
    GetByUserID(ctx context.Context, tx *gorm.DB, userID string) (entities.DesignerProfile, error)
    Update(ctx context.Context, tx *gorm.DB, profile entities.DesignerProfile) (entities.DesignerProfile, error)
    GetAll(ctx context.Context, tx *gorm.DB) ([]entities.DesignerProfile, error)
  }
	
	designerRepository struct {
		db *gorm.DB
	}
)

func NewDesignerProfileRepository(db *gorm.DB) DesignerInterface {
	return &designerRepository{db: db}
}

func (r *designerRepository) GetAll(ctx context.Context, tx *gorm.DB) ([]entities.DesignerProfile, error) {
	if tx==nil{
		tx = r.db
	}

	var designerProfile []entities.DesignerProfile
	err := tx.WithContext(ctx).Preload("User").Find(&designerProfile).Error
	if err != nil {
		return designerProfile, err
	}

	return designerProfile, nil
}

func (r *designerRepository) GetByUserID(ctx context.Context, tx *gorm.DB, userID string) (entities.DesignerProfile, error) {
	if tx==nil {
		tx = r.db
	}

	var profile entities.DesignerProfile
	err := tx.WithContext(ctx).Where("user_id = ?", userID).Take(&profile).Error
	if err != nil {
		return profile, err
	}

	return profile, nil
}

func (r *designerRepository) GetByID(ctx context.Context, tx *gorm.DB, id string) (entities.DesignerProfile, error) {
	if tx==nil {
		tx = r.db
	}

	var profile entities.DesignerProfile
	err := tx.WithContext(ctx).Where("id = ?", id).Take(&profile).Error
	if err != nil {
		return profile, err
	}

	return profile, nil
}

func (r *designerRepository) Update(ctx context.Context, tx *gorm.DB, profile entities.DesignerProfile) (entities.DesignerProfile, error) {
	if tx==nil{
		tx = r.db
	}

	err := tx.WithContext(ctx).Updates(&profile).Error
	if err != nil {
		return entities.DesignerProfile{}, err
	}

	return profile, nil
}