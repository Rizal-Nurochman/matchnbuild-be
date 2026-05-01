package service

import (
	"context"

	"github.com/Caknoooo/go-pagination"
	"github.com/Rizal-Nurochman/matchnbuild/modules/designer/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/designer/query"
	"github.com/Rizal-Nurochman/matchnbuild/modules/designer/repository"
	"gorm.io/gorm"
)

type (
	designerService struct {
		designerRepo repository.DesignerRepository
		db 					 *gorm.DB
	}

	DesignerService interface {
		GetAll(ctx context.Context, filter *query.DesignerFilter) ([]query.Designer, int64, error)
		GetByID(ctx context.Context, id string) (dto.DesignerProfileResponse, error)
		GetMyProfile(ctx context.Context, userID string) (dto.DesignerProfileResponse, error)
		Update(ctx context.Context, userID string, req dto.DesignerProfileUpdateRequest) (dto.DesignerProfileResponse, error)
	}
)

func NewDesignerProfileService(designerRepo repository.DesignerRepository, db *gorm.DB) DesignerService {
	return &designerService{designerRepo: designerRepo, db: db}
}

func (s *designerService) GetAll(ctx context.Context, filter *query.DesignerFilter) ([]query.Designer, int64, error) {
	return pagination.PaginatedQueryWithIncludable[query.Designer](s.db, filter)
}

func (s *designerService) GetByID(ctx context.Context, id string) (dto.DesignerProfileResponse, error) {
	designer, err := s.designerRepo.GetByID(ctx, nil, id)
	if err != nil {
		return dto.DesignerProfileResponse{}, err
	}

	result := dto.DesignerProfileResponse{
		ID: designer.ID.String(),
		UserID: designer.UserID.String(),
		Name: designer.User.Name,
		ProfilePicture: designer.User.ProfilePicture,
		Bio: designer.Bio,
		ExperienceYears: designer.ExperienceYears,
		IsVerified: designer.IsVerified,
		IsAvailable: designer.IsAvailable,
		Location: designer.Location,
	}

	return result, nil
}

func (s *designerService) GetMyProfile(ctx context.Context, userID string) (dto.DesignerProfileResponse, error) {
	profile, err := s.designerRepo.GetByUserID(ctx, nil, userID)
	if err != nil {
		return dto.DesignerProfileResponse{}, dto.ErrDesignerProfileNotFound
	}

	result := dto.DesignerProfileResponse{
		ID:              profile.ID.String(),
		UserID:          profile.UserID.String(),
		Name:            profile.User.Name,
		ProfilePicture:  profile.User.ProfilePicture,
		Bio:             profile.Bio,
		ExperienceYears: profile.ExperienceYears,
		IsVerified:      profile.IsVerified,
		IsAvailable:     profile.IsAvailable,
		Location:        profile.Location,
	}

	return result, nil
}

func (s *designerService) Update(ctx context.Context, userID string, req dto.DesignerProfileUpdateRequest) (dto.DesignerProfileResponse, error) {
	profile, err := s.designerRepo.GetByUserID(ctx, nil, userID)
	if err != nil {
		return dto.DesignerProfileResponse{}, dto.ErrDesignerProfileNotFound
	}

	if req.Bio != nil {
		profile.Bio = *req.Bio
	}
	if req.Location != nil {
		profile.Location = *req.Location
	}
	if req.ExperienceYears != nil {
		profile.ExperienceYears = *req.ExperienceYears
	}
	if req.IsAvailable != nil {
		profile.IsAvailable = *req.IsAvailable
	}
	if req.BankAccountNumber != nil {
		profile.BankAccountNumber = *req.BankAccountNumber
	}

	updated, err := s.designerRepo.Update(ctx, nil, profile)
	if err != nil {
		return dto.DesignerProfileResponse{}, err
	}

	result := dto.DesignerProfileResponse{
		ID:              updated.ID.String(),
		UserID:          updated.UserID.String(),
		Name:            profile.User.Name,
		ProfilePicture:  profile.User.ProfilePicture,
		Bio:             updated.Bio,
		ExperienceYears: updated.ExperienceYears,
		IsVerified:      updated.IsVerified,
		IsAvailable:     updated.IsAvailable,
		Location:        updated.Location,
	}

	return result, nil
}
