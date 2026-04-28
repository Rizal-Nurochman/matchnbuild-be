package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

type (
	userPreferenceRepository struct {
		db *gorm.DB
	}

	UserPreferenceRepository interface {
		GetByUserID(ctx context.Context, tx *gorm.DB, userID string) (entities.UserPreference, error)
		Create(ctx context.Context, tx *gorm.DB, preference entities.UserPreference) (entities.UserPreference, error)
		Update(ctx context.Context, tx *gorm.DB, preference entities.UserPreference) (entities.UserPreference, error)
	}
)

func NewUserPreferenceRepository(db *gorm.DB) UserPreferenceRepository  {
	return &userPreferenceRepository{db: db}
}

func (r *userPreferenceRepository) GetByUserID(ctx context.Context, tx *gorm.DB, userID string) (entities.UserPreference, error) {
	if tx == nil {
		tx = r.db
	}

	var preference entities.UserPreference
	err := tx.WithContext(ctx).Where("user_id = ?", userID).Take(&preference).Error
	if err != nil {
		return preference, err
	}

	return preference, nil
}

func (r *userPreferenceRepository) Create(ctx context.Context, tx *gorm.DB, preference entities.UserPreference) (entities.UserPreference, error) {
	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Create(&preference).Error
	if err != nil {
		return preference, err
	}

	return preference, nil
}

func (r *userPreferenceRepository) Update(ctx context.Context, tx *gorm.DB, preference entities.UserPreference) (entities.UserPreference, error) {
	if tx == nil {
		tx = r.db
	}

	err := tx.WithContext(ctx).Updates(&preference).Error
	if err != nil {
		return preference, err
	}

	return preference, nil
}