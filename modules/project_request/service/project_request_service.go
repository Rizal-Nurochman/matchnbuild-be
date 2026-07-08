package service

import (
	"context"
	"fmt"
	"time"

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
	GetIncomingByUserID(ctx context.Context, userID string) ([]dto.ProjectRequestResponse, error)
	MarkAsDone(ctx context.Context, id string, designerUserID string) (dto.ProjectRequestResponse, error)
	MarkAsCompleted(ctx context.Context, id string, clientID string) (dto.ProjectRequestResponse, error)
	MarkAsRevision(ctx context.Context, id string, clientID string) (dto.ProjectRequestResponse, error)
	MarkAsInProgress(ctx context.Context, id string, designerUserID string) (dto.ProjectRequestResponse, error)
}

type projectRequestService struct {
	projectRequestRepo   repository.ProjectRequestRepository
	conversationRepo     repository.ConversationRepository
	designerProfileRepo  repository.DesignerProfileRepository
	conversationParticipantRepo repository.ConversationParticipantRepository
	db                   *gorm.DB
}

func NewProjectRequestService(
	prRepo repository.ProjectRequestRepository,
	convRepo repository.ConversationRepository,
	dpRepo repository.DesignerProfileRepository,
	participantRepo repository.ConversationParticipantRepository,
	db *gorm.DB,
) ProjectRequestService {
	return &projectRequestService{
		projectRequestRepo:  prRepo,
		conversationRepo:    convRepo,
		designerProfileRepo: dpRepo,
		conversationParticipantRepo: participantRepo,
		db:                  db,
	}
}

func (s *projectRequestService) Create(ctx context.Context, req dto.ProjectRequestCreateRequest, clientID string) (dto.ProjectRequestResponse, error) {
	designerProfile, err := s.designerProfileRepo.GetByID(ctx, s.db, req.DesignerID)
	if err != nil {
		return dto.ProjectRequestResponse{}, dto.ErrDesignerProfileNotFound
	}

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

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return dto.ProjectRequestResponse{}, tx.Error
	}

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

	conversation := entities.Conversation{
		ID:               uuid.New(),
		ProjectRequestID: createdPR.ID,
	}

	createdConv, err := s.conversationRepo.Create(ctx, tx, conversation)
	if err != nil {
		tx.Rollback()
		return dto.ProjectRequestResponse{}, err
	}

	participants := []entities.ConversationParticipant{
		{
			ConversationID: createdConv.ID,
			UserID:         clientUUID,
			Role:           constants.ENUM_ROLE_CLIENT,
		},
		{
			ConversationID: createdConv.ID,
			UserID:         designerProfile.UserID,
			Role:           constants.ENUM_ROLE_DESIGNER,
		},
	}

	if err := s.conversationParticipantRepo.CreateMany(ctx, tx, participants); err != nil {
		tx.Rollback()
		return dto.ProjectRequestResponse{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return dto.ProjectRequestResponse{}, err
	}

	fullPR, err := s.projectRequestRepo.GetByID(ctx, s.db, createdPR.ID.String())
	if err != nil {
		return dto.ProjectRequestResponse{}, err
	}

	response := toProjectRequestResponse(fullPR)
	response.ConversationID = createdConv.ID.String()

	return response, nil
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

func (s *projectRequestService) GetIncomingByUserID(ctx context.Context, userID string) ([]dto.ProjectRequestResponse, error) {
	designerProfile, err := s.designerProfileRepo.GetByUserID(ctx, s.db, userID)
	if err != nil {
		return nil, dto.ErrDesignerProfileNotFound
	}

	prs, err := s.projectRequestRepo.GetByDesignerID(ctx, s.db, designerProfile.ID.String())
	if err != nil {
		return nil, err
	}

	return toProjectRequestResponses(prs), nil
}

func toProjectRequestResponse(pr entities.ProjectRequest) dto.ProjectRequestResponse {
	quotations := make([]dto.QuotationInfo, 0, len(pr.Quotations))
	for _, q := range pr.Quotations {
		qi := dto.QuotationInfo{
			ID:           q.ID.String(),
			ScopeOfWork:  q.ScopeOfWork,
			OfferedPrice: q.OfferedPrice.InexactFloat64(),
			DurationDays: q.DurationDays,
			Status:       q.Status,
		}
		if q.Order != nil {
			qi.Order = &dto.OrderInfo{
				ID:            q.Order.ID.String(),
				PaymentStatus: q.Order.PaymentStatus,
				WorkStatus:    q.Order.WorkStatus,
			}
		}
		quotations = append(quotations, qi)
	}

	return dto.ProjectRequestResponse{
		ID: pr.ID.String(),
		Client: dto.ClientInfo{
			ClientID: pr.ClientID.String(),
			Name:     pr.Client.Name,
		},
		Designer: dto.DesignerInfo{
			DesignerID: pr.DesignerID.String(),
			Name:       pr.Designer.User.Name,
		},
		Description:      pr.Description,
		InitialBudget:    pr.InitialBudget.InexactFloat64(),
		AreaSize:         pr.AreaSize,
		LocationPhotoURL: pr.LocationPhotoURL,
		LayoutSketchURL:  pr.LayoutSketchURL,
		Status:           pr.Status,
		Quotations:       quotations,
	}
}

func (s *projectRequestService) MarkAsDone(ctx context.Context, id string, designerUserID string) (dto.ProjectRequestResponse, error) {
	pr, err := s.projectRequestRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return dto.ProjectRequestResponse{}, dto.ErrProjectRequestNotFound
	}

	designerProfile, err := s.designerProfileRepo.GetByUserID(ctx, s.db, designerUserID)
	if err != nil {
		return dto.ProjectRequestResponse{}, dto.ErrDesignerProfileNotFound
	}
	if pr.DesignerID.String() != designerProfile.ID.String() {
		return dto.ProjectRequestResponse{}, dto.ErrNotProjectRequestDesigner
	}
	if pr.Status != constants.PROJECT_REQUEST_STATUS_IN_PROGRESS {
		return dto.ProjectRequestResponse{}, dto.ErrInvalidStatusTransition
	}

	if err := s.projectRequestRepo.UpdateStatus(ctx, s.db, id, constants.PROJECT_REQUEST_STATUS_DONE); err != nil {
		return dto.ProjectRequestResponse{}, err
	}

	updated, err := s.projectRequestRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return dto.ProjectRequestResponse{}, err
	}
	return toProjectRequestResponse(updated), nil
}

func (s *projectRequestService) MarkAsCompleted(ctx context.Context, id string, clientID string) (dto.ProjectRequestResponse, error) {
	pr, err := s.projectRequestRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return dto.ProjectRequestResponse{}, dto.ErrProjectRequestNotFound
	}
	if pr.ClientID.String() != clientID {
		return dto.ProjectRequestResponse{}, dto.ErrNotProjectRequestClient
	}
	if pr.Status != constants.PROJECT_REQUEST_STATUS_DONE {
		return dto.ProjectRequestResponse{}, dto.ErrInvalidStatusTransition
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return dto.ProjectRequestResponse{}, tx.Error
	}

	// Update order.work_status → Completed untuk quotation yang accepted
	for _, q := range pr.Quotations {
		if q.Status == constants.QUOTATION_STATUS_ACCEPTED {
			now := time.Now()
			if err := tx.Model(&entities.Order{}).
				Where("quotation_id = ?", q.ID).
				Updates(map[string]interface{}{
					"work_status":  constants.ORDER_WORK_STATUS_COMPLETED,
					"completed_at": now,
				}).Error; err != nil {
				tx.Rollback()
				return dto.ProjectRequestResponse{}, err
			}
			break
		}
	}

	if err := s.projectRequestRepo.UpdateStatus(ctx, tx, id, constants.PROJECT_REQUEST_STATUS_COMPLETED); err != nil {
		tx.Rollback()
		return dto.ProjectRequestResponse{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return dto.ProjectRequestResponse{}, err
	}

	updated, err := s.projectRequestRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return dto.ProjectRequestResponse{}, err
	}
	return toProjectRequestResponse(updated), nil
}

func (s *projectRequestService) MarkAsRevision(ctx context.Context, id string, clientID string) (dto.ProjectRequestResponse, error) {
	pr, err := s.projectRequestRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return dto.ProjectRequestResponse{}, dto.ErrProjectRequestNotFound
	}
	if pr.ClientID.String() != clientID {
		return dto.ProjectRequestResponse{}, dto.ErrNotProjectRequestClient
	}
	if pr.Status != constants.PROJECT_REQUEST_STATUS_DONE {
		return dto.ProjectRequestResponse{}, dto.ErrInvalidStatusTransition
	}

	if err := s.projectRequestRepo.UpdateStatus(ctx, s.db, id, constants.PROJECT_REQUEST_STATUS_REVISION); err != nil {
		return dto.ProjectRequestResponse{}, err
	}

	updated, err := s.projectRequestRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return dto.ProjectRequestResponse{}, err
	}
	return toProjectRequestResponse(updated), nil
}

func (s *projectRequestService) MarkAsInProgress(ctx context.Context, id string, designerUserID string) (dto.ProjectRequestResponse, error) {
	pr, err := s.projectRequestRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return dto.ProjectRequestResponse{}, dto.ErrProjectRequestNotFound
	}

	designerProfile, err := s.designerProfileRepo.GetByUserID(ctx, s.db, designerUserID)
	if err != nil {
		return dto.ProjectRequestResponse{}, dto.ErrDesignerProfileNotFound
	}
	if pr.DesignerID.String() != designerProfile.ID.String() {
		return dto.ProjectRequestResponse{}, dto.ErrNotProjectRequestDesigner
	}
	if pr.Status != constants.PROJECT_REQUEST_STATUS_REVISION {
		return dto.ProjectRequestResponse{}, dto.ErrInvalidStatusTransition
	}

	if err := s.projectRequestRepo.UpdateStatus(ctx, s.db, id, constants.PROJECT_REQUEST_STATUS_IN_PROGRESS); err != nil {
		return dto.ProjectRequestResponse{}, err
	}

	updated, err := s.projectRequestRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return dto.ProjectRequestResponse{}, err
	}
	return toProjectRequestResponse(updated), nil
}

func toProjectRequestResponses(prs []entities.ProjectRequest) []dto.ProjectRequestResponse {
	responses := make([]dto.ProjectRequestResponse, len(prs))
	for i, pr := range prs {
		responses[i] = toProjectRequestResponse(pr)
	}
	return responses
}
