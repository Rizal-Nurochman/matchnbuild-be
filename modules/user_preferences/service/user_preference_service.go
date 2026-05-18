package service

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"github.com/Rizal-Nurochman/matchnbuild/modules/user_preferences/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/user_preferences/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type (
	userPreferenceService struct {
		userPreferenceRepository repository.UserPreferenceRepository
	}

	UserPreferenceService interface {
		GetByUserID(ctx context.Context, userID string) (dto.PreferenceResponse, error)
		Create(ctx context.Context, userID string, req dto.CreatePreferenceRequest) (dto.PreferenceResponse, error)
		Update(ctx context.Context, userID string, req dto.UpdatePreferenceRequest) (dto.PreferenceResponse, error)
	}
)

func NewUserPreferenceService(userPreferenceRepo repository.UserPreferenceRepository) UserPreferenceService {
	return &userPreferenceService{userPreferenceRepo}
}

func (s *userPreferenceService) GetByUserID(ctx context.Context, userID string) (dto.PreferenceResponse, error) {
	preference, err := s.userPreferenceRepository.GetByUserID(ctx, nil, userID)
	if err != nil {
		return dto.PreferenceResponse{}, dto.ErrPreferenceNotFound
	}

	return dto.ToUserPreferenceResponse(preference), nil
}

func (s *userPreferenceService) Create(ctx context.Context, userID string, req dto.CreatePreferenceRequest) (dto.PreferenceResponse, error) {
	_, err := s.userPreferenceRepository.GetByUserID(ctx, nil, userID)
	if err == nil {
		return dto.PreferenceResponse{}, dto.ErrPreferenceAlreadyExist
	}

	prefenceItem := entities.UserPreference{
		ID:                uuid.New(),
		UserID:            uuid.MustParse(userID),
		PreferredStyle:    req.PreferredStyle,
		BudgetMin:         decimal.NewFromFloat(req.BudgetMin),
		BudgetMax:         decimal.NewFromFloat(req.BudgetMax),
		PreferredLocation: req.PreferredLocation,
		IsOnboarded:       true,
	}

	created, err := s.userPreferenceRepository.Create(ctx, nil, prefenceItem)
	if err != nil {
		return dto.PreferenceResponse{}, err
	}

	return dto.ToUserPreferenceResponse(created), nil
}

func (s *userPreferenceService) Update(ctx context.Context, userID string, req dto.UpdatePreferenceRequest) (dto.PreferenceResponse, error) {
	preference, err := s.userPreferenceRepository.GetByUserID(ctx, nil, userID)
	if err != nil {
		return dto.PreferenceResponse{}, dto.ErrPreferenceNotFound
	}
	if req.PreferredStyle != nil {
		preference.PreferredStyle = *req.PreferredStyle
	}
	if req.BudgetMin != nil {
		preference.BudgetMin = decimal.NewFromFloat(*req.BudgetMin)
	}
	if req.BudgetMax != nil {
		preference.BudgetMax = decimal.NewFromFloat(*req.BudgetMax)
	}
	if req.PreferredLocation != nil {
		preference.PreferredLocation = *req.PreferredLocation
	}
	updated, err := s.userPreferenceRepository.Update(ctx, nil, preference)
	if err != nil {
		return dto.PreferenceResponse{}, err
	}

	return dto.ToUserPreferenceResponse(updated), nil
}