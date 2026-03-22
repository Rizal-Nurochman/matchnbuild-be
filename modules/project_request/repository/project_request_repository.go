package repository

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

type (
	ProjectRequestRepository interface {
		Create(ctx context.Context, tx *gorm.DB, projectRequest entities.ProjectRequest) (entities.ProjectRequest, error)
		GetByID(ctx context.Context, tx *gorm.DB, id string) (entities.ProjectRequest, error)
		GetByClientID(ctx context.Context, tx *gorm.DB, clientID string) ([]entities.ProjectRequest, error)
		GetByDesignerID(ctx context.Context, tx *gorm.DB, designerID string) ([]entities.ProjectRequest, error)
		Update(ctx context.Context, tx *gorm.DB, projectRequest entities.ProjectRequest) (entities.ProjectRequest, error)
	}

	projectRequestRepository struct {
		db *gorm.DB
	}
)

func NewProjectRequestRepository(db *gorm.DB) ProjectRequestRepository {
	return &projectRequestRepository{db: db}
}

func (r *projectRequestRepository) Create(ctx context.Context, tx *gorm.DB, projectRequest entities.ProjectRequest) (entities.ProjectRequest, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&projectRequest).Error; err != nil {
		return entities.ProjectRequest{}, err
	}

	return projectRequest, nil
}

func (r *projectRequestRepository) GetByID(ctx context.Context, tx *gorm.DB, id string) (entities.ProjectRequest, error) {
	if tx == nil {
		tx = r.db
	}

	var projectRequest entities.ProjectRequest
	if err := tx.WithContext(ctx).Where("id = ?", id).Take(&projectRequest).Error; err != nil {
		return entities.ProjectRequest{}, err
	}

	return projectRequest, nil
}

func (r *projectRequestRepository) GetByClientID(ctx context.Context, tx *gorm.DB, clientID string) ([]entities.ProjectRequest, error) {
	if tx == nil {
		tx = r.db
	}

	var projectRequests []entities.ProjectRequest
	if err := tx.WithContext(ctx).Where("client_id = ?", clientID).Find(&projectRequests).Error; err != nil {
		return nil, err
	}

	return projectRequests, nil
}

func (r *projectRequestRepository) GetByDesignerID(ctx context.Context, tx *gorm.DB, designerID string) ([]entities.ProjectRequest, error) {
	if tx == nil {
		tx = r.db
	}

	var projectRequests []entities.ProjectRequest
	if err := tx.WithContext(ctx).Where("designer_id = ?", designerID).Find(&projectRequests).Error; err != nil {
		return nil, err
	}

	return projectRequests, nil
}

func (r *projectRequestRepository) Update(ctx context.Context, tx *gorm.DB, projectRequest entities.ProjectRequest) (entities.ProjectRequest, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Updates(&projectRequest).Error; err != nil {
		return entities.ProjectRequest{}, err
	}

	return projectRequest, nil
}
