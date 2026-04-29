package service

import (
	"context"
	"errors"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"github.com/Rizal-Nurochman/matchnbuild/modules/design_item/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/design_item/repository"
	designerRepo "github.com/Rizal-Nurochman/matchnbuild/modules/designer/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type (
	designItemService struct {
		designItemRepo repository.DesignItemRepository
		designerRepo   designerRepo.DesignerRepository
		db             *gorm.DB
	}

	DesignItemService interface {
		Create(ctx context.Context, userID string, req dto.DesignItemCreateRequest) (dto.DesignItemResponse, error)
		GetAll(ctx context.Context) ([]dto.DesignItemResponse, error)
		GetAllFeatures(ctx context.Context, category string) ([]dto.FeatureResponse, error)
		GetByID(ctx context.Context, id string) (dto.DesignItemResponse, error)
		GetMyItems(ctx context.Context, userID string) ([]dto.DesignItemResponse, error)
		Update(ctx context.Context, userID string, designItemID string, req dto.DesignItemUpdateRequest) (dto.DesignItemResponse, error)
		Delete(ctx context.Context, userID string, designItemID string) error
	}
)

func NewDesignItemService(
	designItemRepo repository.DesignItemRepository,
	designerRepo designerRepo.DesignerRepository,
	db *gorm.DB,
) DesignItemService {
	return &designItemService{
		designItemRepo: designItemRepo,
		designerRepo:   designerRepo,
		db:             db,
	}
}

func (s *designItemService) Create(ctx context.Context, userID string, req dto.DesignItemCreateRequest) (dto.DesignItemResponse, error) {
	designer, err := s.designerRepo.GetByUserID(ctx, nil, userID)
	if err != nil {
		return dto.DesignItemResponse{}, dto.ErrDesignItemNotFound
	}

	switch req.Category {
	case "architecture":
		if req.LandAreaMin == nil || req.LandAreaMax == nil ||
			req.BuildingArea == nil || req.NumFloors == nil {
			return dto.DesignItemResponse{}, errors.New("Field arsitektur harus diisi!")
		}
	case "interior":
		if req.RoomType == nil || req.RoomArea == nil {
			return dto.DesignItemResponse{}, errors.New("Field interior harus diisi!")
		}
	default:
		return dto.DesignItemResponse{}, errors.New("Kategori harus 'Arsitektur' atau 'Interior'")
	}

	var features []entities.DesignItemFeature
	for _, fID := range req.FeatureIDs {
		featureUUID, err := uuid.Parse(fID)
		if err != nil {
			return dto.DesignItemResponse{}, err
		}
		features = append(features, entities.DesignItemFeature{
			FeatureID: featureUUID,
		})
	}

	designItem := entities.DesignItem{
		ID:              uuid.New(),
		DesignerID:      designer.ID,
		Title:           req.Title,
		Description:     req.Description,
		Style:           req.Style,
		Category:        req.Category,
		LandAreaMin:     req.LandAreaMin,
		LandAreaMax:     req.LandAreaMax,
		BuildingArea:    req.BuildingArea,
		NumFloors:       req.NumFloors,
		NumBedrooms:     req.NumBedrooms,
		RoomType:        req.RoomType,
		RoomArea:        req.RoomArea,
		EstimatedBudget: decimal.NewFromFloat(req.EstimatedBudget),
		PriceStartFrom:  decimal.NewFromFloat(req.PriceStartFrom),
		ImageURL:        req.ImageURL,
		Features:        features,
	}

	created, err := s.designItemRepo.Create(ctx, nil, designItem)
	if err != nil {
		return dto.DesignItemResponse{}, err
	}

	result, err := s.designItemRepo.GetByID(ctx, nil, created.ID.String())
	if err != nil {
		return dto.DesignItemResponse{}, err
	}

	return dto.ToDesignItemResponse(result), nil
}

func (s *designItemService) GetAll(ctx context.Context) ([]dto.DesignItemResponse, error) {
	items, err := s.designItemRepo.GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	var result []dto.DesignItemResponse
	for _, item := range items {
		result = append(result, dto.ToDesignItemResponse(item))
	}

	return result, nil
}

func (s *designItemService) GetAllFeatures(ctx context.Context, category string) ([]dto.FeatureResponse, error) {
	var features []entities.Feature
	var err error

	if category != "" {
		features, err = s.designItemRepo.GetByCategory(ctx, nil, category)
	} else {
		features, err = s.designItemRepo.GetAllFeatures(ctx, nil)
	}

	if err != nil {
		return nil, err
	}

	var result []dto.FeatureResponse
	for _, data := range features {
		result = append(result, dto.FeatureResponse{
			ID:   data.ID.String(),
			Name: data.Name,
		})
	}

	return result, nil
}

func (s *designItemService) GetByID(ctx context.Context, id string) (dto.DesignItemResponse, error) {
	item, err := s.designItemRepo.GetByID(ctx, nil, id)
	if err != nil {
		return dto.DesignItemResponse{}, dto.ErrDesignItemNotFound
	}

	return dto.ToDesignItemResponse(item), nil
}

func (s *designItemService) GetMyItems(ctx context.Context, userID string) ([]dto.DesignItemResponse, error) {
	designer, err := s.designerRepo.GetByUserID(ctx, nil, userID)
	if err != nil {
		return nil, dto.ErrDesignItemNotFound
	}

	items, err := s.designItemRepo.GetByDesignerID(ctx, nil, designer.ID.String())
	if err != nil {
		return nil, err
	}

	var result []dto.DesignItemResponse
	for _, item := range items {
		result = append(result, dto.ToDesignItemResponse(item))
	}

	return result, nil
}

func (s *designItemService) Update(ctx context.Context, userID string, designItemID string, req dto.DesignItemUpdateRequest) (dto.DesignItemResponse, error) {
	designer, err := s.designerRepo.GetByUserID(ctx, nil, userID)
	if err != nil {
		return dto.DesignItemResponse{}, dto.ErrDesignItemNotFound
	}

	item, err := s.designItemRepo.GetByID(ctx, nil, designItemID)
	if err != nil {
		return dto.DesignItemResponse{}, dto.ErrDesignItemNotFound
	}

	if item.DesignerID != designer.ID {
		return dto.DesignItemResponse{}, dto.ErrNotDesignItemOwner
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return dto.DesignItemResponse{}, tx.Error
	}
	defer tx.Rollback()

	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Style != nil {
		item.Style = *req.Style
	}
	if req.Category != nil {
		item.Category = *req.Category
	}
	if req.LandAreaMin != nil {
		item.LandAreaMin = req.LandAreaMin
	}
	if req.LandAreaMax != nil {
		item.LandAreaMax = req.LandAreaMax
	}
	if req.BuildingArea != nil {
		item.BuildingArea = req.BuildingArea
	}
	if req.NumFloors != nil {
		item.NumFloors = req.NumFloors
	}
	if req.NumBedrooms != nil {
		item.NumBedrooms = req.NumBedrooms
	}
	if req.RoomType != nil {
		item.RoomType = req.RoomType
	}
	if req.RoomArea != nil {
		item.RoomArea = req.RoomArea
	}
	if req.EstimatedBudget != nil {
		item.EstimatedBudget = decimal.NewFromFloat(*req.EstimatedBudget)
	}
	if req.PriceStartFrom != nil {
		item.PriceStartFrom = decimal.NewFromFloat(*req.PriceStartFrom)
	}
	if req.ImageURL != "" {
		item.ImageURL = req.ImageURL
	}

	if err := tx.Updates(&item).Error; err != nil {
		return dto.DesignItemResponse{}, err
	}

	if req.FeatureIDs != nil {
		if err := tx.Where("design_item_id = ?", item.ID).
			Delete(&entities.DesignItemFeature{}).Error; err != nil {
			return dto.DesignItemResponse{}, err
		}

		for _, fID := range req.FeatureIDs {
			featureUUID, err := uuid.Parse(fID)
			if err != nil {
				return dto.DesignItemResponse{}, err
			}

			if err := tx.Create(&entities.DesignItemFeature{
				DesignItemID: item.ID,
				FeatureID:    featureUUID,
			}).Error; err != nil {
				return dto.DesignItemResponse{}, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return dto.DesignItemResponse{}, err
	}

	result, err := s.designItemRepo.GetByID(ctx, nil, item.ID.String())
	if err != nil {
		return dto.DesignItemResponse{}, err
	}

	return dto.ToDesignItemResponse(result), nil
}

func (s *designItemService) Delete(ctx context.Context, userID string, designItemID string) error {
	designer, err := s.designerRepo.GetByUserID(ctx, nil, userID)
	if err != nil {
		return dto.ErrDesignItemNotFound
	}

	item, err := s.designItemRepo.GetByID(ctx, nil, designItemID)
	if err != nil {
		return dto.ErrDesignItemNotFound
	}

	if item.DesignerID != designer.ID {
		return dto.ErrNotDesignItemOwner
	}

	return s.designItemRepo.Delete(ctx, nil, designItemID)
}
