package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

type (
	DesignerProfileRepository interface {
		GetByID(ctx context.Context, tx *gorm.DB, id string) (entities.DesignerProfile, error)
		GetByUserID(ctx context.Context, tx *gorm.DB, userID string) (entities.DesignerProfile, error)
	}

	designerProfileRepository struct {
		db *gorm.DB
	}
)

func NewDesignerProfileRepository(db *gorm.DB) DesignerProfileRepository {
	return &designerProfileRepository{db: db}
}

func (r *designerProfileRepository) GetByID(ctx context.Context, tx *gorm.DB, id string) (entities.DesignerProfile, error) {
	if tx == nil {
		tx = r.db
	}

	var profile entities.DesignerProfile
	if err := tx.WithContext(ctx).Where("id = ?", id).Take(&profile).Error; err != nil {
		return entities.DesignerProfile{}, err
	}

	return profile, nil
}

func (r *designerProfileRepository) GetByUserID(ctx context.Context, tx *gorm.DB, userID string) (entities.DesignerProfile, error) {
	if tx == nil {
		tx = r.db
	}

	var profile entities.DesignerProfile
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).Take(&profile).Error; err != nil {
		return entities.DesignerProfile{}, err
	}

	return profile, nil
}
