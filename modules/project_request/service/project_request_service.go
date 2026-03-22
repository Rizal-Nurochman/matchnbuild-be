package service

import (
	"context"
	"fmt"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"github.com/Rizal-Nurochman/matchnbuild/modules/project_request/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/project_request/repository"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProjectRequestService interface {
	Create(ctx context.Context, req dto.ProjectRequestCreateRequest, clientID string) (dto.ProjectRequestResponse, error)
	GetByID(ctx context.Context, id string) (dto.ProjectRequestResponse, error)
	GetByClientID(ctx context.Context, clientID string) ([]dto.ProjectRequestResponse, error)
	GetByDesignerID(ctx context.Context, designerID string) ([]dto.ProjectRequestResponse, error)
}

type projectRequestService struct {
	projectRequestRepo   repository.ProjectRequestRepository
	conversationRepo     repository.ConversationRepository
	designerProfileRepo  repository.DesignerProfileRepository
	db                   *gorm.DB
}

func NewProjectRequestService(
	prRepo repository.ProjectRequestRepository,
	convRepo repository.ConversationRepository,
	dpRepo repository.DesignerProfileRepository,
	db *gorm.DB,
) ProjectRequestService {
	return &projectRequestService{
		projectRequestRepo:  prRepo,
		conversationRepo:    convRepo,
		designerProfileRepo: dpRepo,
		db:                  db,
	}
}

func (s *projectRequestService) Create(ctx context.Context, req dto.ProjectRequestCreateRequest, clientID string) (dto.ProjectRequestResponse, error) {
	// Validate designer exists
	designerProfile, err := s.designerProfileRepo.GetByID(ctx, s.db, req.DesignerID)
	if err != nil {
		return dto.ProjectRequestResponse{}, dto.ErrDesignerProfileNotFound
	}

	// Prevent self-request
	if designerProfile.UserID.String() == clientID {
		return dto.ProjectRequestResponse{}, dto.ErrCannotRequestOwnDesign
	}

	clientUUID, err := uuid.Parse(clientID)
	if err != nil {
		return dto.ProjectRequestResponse{}, fmt.Errorf("invalid client id: %w", err)
	}

	designerUUID, err := uuid.Parse(req.DesignerID)
	if err != nil {
		return dto.ProjectRequestResponse{}, fmt.Errorf("invalid designer id: %w", err)
	}

	// Start transaction: create ProjectRequest + Conversation
	tx := s.db.Begin()

	projectRequest := entities.ProjectRequest{
		ID:               uuid.New(),
		ClientID:         clientUUID,
		DesignerID:       designerUUID,
		Description:      req.Description,
		InitialBudget:    decimal.NewFromFloat(req.InitialBudget),
		AreaSize:         req.AreaSize,
		LocationPhotoURL: req.LocationPhotoURL,
		LayoutSketchURL:  req.LayoutSketchURL,
		Status:           constants.PROJECT_REQUEST_STATUS_OPEN,
	}

	createdPR, err := s.projectRequestRepo.Create(ctx, tx, projectRequest)
	if err != nil {
		tx.Rollback()
		return dto.ProjectRequestResponse{}, err
	}

	// Auto-create Conversation
	conversation := entities.Conversation{
		ID:               uuid.New(),
		ProjectRequestID: createdPR.ID,
	}

	createdConv, err := s.conversationRepo.Create(ctx, tx, conversation)
	if err != nil {
		tx.Rollback()
		return dto.ProjectRequestResponse{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return dto.ProjectRequestResponse{}, err
	}

	return dto.ProjectRequestResponse{
		ID:               createdPR.ID.String(),
		ClientID:         createdPR.ClientID.String(),
		DesignerID:       createdPR.DesignerID.String(),
		Description:      createdPR.Description,
		InitialBudget:    createdPR.InitialBudget.InexactFloat64(),
		AreaSize:         createdPR.AreaSize,
		LocationPhotoURL: createdPR.LocationPhotoURL,
		LayoutSketchURL:  createdPR.LayoutSketchURL,
		Status:           createdPR.Status,
		ConversationID:   createdConv.ID.String(),
	}, nil
}

func (s *projectRequestService) GetByID(ctx context.Context, id string) (dto.ProjectRequestResponse, error) {
	pr, err := s.projectRequestRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return dto.ProjectRequestResponse{}, dto.ErrProjectRequestNotFound
	}

	return toProjectRequestResponse(pr), nil
}

func (s *projectRequestService) GetByClientID(ctx context.Context, clientID string) ([]dto.ProjectRequestResponse, error) {
	prs, err := s.projectRequestRepo.GetByClientID(ctx, s.db, clientID)
	if err != nil {
		return nil, err
	}

	return toProjectRequestResponses(prs), nil
}

func (s *projectRequestService) GetByDesignerID(ctx context.Context, designerID string) ([]dto.ProjectRequestResponse, error) {
	prs, err := s.projectRequestRepo.GetByDesignerID(ctx, s.db, designerID)
	if err != nil {
		return nil, err
	}

	return toProjectRequestResponses(prs), nil
}

func toProjectRequestResponse(pr entities.ProjectRequest) dto.ProjectRequestResponse {
	return dto.ProjectRequestResponse{
		ID:               pr.ID.String(),
		ClientID:         pr.ClientID.String(),
		DesignerID:       pr.DesignerID.String(),
		Description:      pr.Description,
		InitialBudget:    pr.InitialBudget.InexactFloat64(),
		AreaSize:         pr.AreaSize,
		LocationPhotoURL: pr.LocationPhotoURL,
		LayoutSketchURL:  pr.LayoutSketchURL,
		Status:           pr.Status,
	}
}

func toProjectRequestResponses(prs []entities.ProjectRequest) []dto.ProjectRequestResponse {
	responses := make([]dto.ProjectRequestResponse, len(prs))
	for i, pr := range prs {
		responses[i] = toProjectRequestResponse(pr)
	}
	return responses
}
