package service

import (
	"context"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"github.com/Rizal-Nurochman/matchnbuild/modules/project_request/repository"
	"github.com/Rizal-Nurochman/matchnbuild/modules/quotation/dto"
	quotationRepo "github.com/Rizal-Nurochman/matchnbuild/modules/quotation/repository"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type QuotationService interface {
	Create(ctx context.Context, req dto.QuotationCreateRequest, designerID string) (dto.QuotationResponse, error)
	GetByID(ctx context.Context, id string) (dto.QuotationResponse, error)
	Accept(ctx context.Context, quotationID string, clientID string) (dto.QuotationAcceptResponse, error)
	Reject(ctx context.Context, quotationID string, clientID string) error
}

type quotationService struct {
	quotationRepo      quotationRepo.QuotationRepository
	orderRepo          quotationRepo.OrderRepository
	projectRequestRepo repository.ProjectRequestRepository
	designerProfileRepo repository.DesignerProfileRepository
	db                 *gorm.DB
}

func NewQuotationService(
	qRepo quotationRepo.QuotationRepository,
	oRepo quotationRepo.OrderRepository,
	prRepo repository.ProjectRequestRepository,
	dpRepo repository.DesignerProfileRepository,
	db *gorm.DB,
) QuotationService {
	return &quotationService{
		quotationRepo:       qRepo,
		orderRepo:           oRepo,
		projectRequestRepo:  prRepo,
		designerProfileRepo: dpRepo,
		db:                  db,
	}
}

func (s *quotationService) Create(ctx context.Context, req dto.QuotationCreateRequest, designerID string) (dto.QuotationResponse, error) {
	// Validate project request exists and is open
	pr, err := s.projectRequestRepo.GetByID(ctx, s.db, req.ProjectRequestID)
	if err != nil {
		return dto.QuotationResponse{}, dto.ErrQuotationNotFound
	}

	if pr.Status != constants.PROJECT_REQUEST_STATUS_OPEN {
		return dto.QuotationResponse{}, dto.ErrProjectRequestNotOpen
	}

	// Validate the designer is the target of this request
	designerProfile, err := s.designerProfileRepo.GetByUserID(ctx, s.db, designerID)
	if err != nil {
		return dto.QuotationResponse{}, dto.ErrNotDesignerOfRequest
	}

	if pr.DesignerID != designerProfile.ID {
		return dto.QuotationResponse{}, dto.ErrNotDesignerOfRequest
	}

	// Check if quotation already exists
	_, err = s.quotationRepo.GetByProjectRequestID(ctx, s.db, req.ProjectRequestID)
	if err == nil {
		return dto.QuotationResponse{}, dto.ErrQuotationAlreadyExists
	}

	quotation := entities.Quotation{
		ID:               uuid.New(),
		ProjectRequestID: pr.ID,
		DesignerID:       designerProfile.ID,
		ScopeOfWork:      req.ScopeOfWork,
		OfferedPrice:     decimal.NewFromFloat(req.OfferedPrice),
		DurationDays:     req.DurationDays,
		Status:           constants.QUOTATION_STATUS_PENDING,
	}

	created, err := s.quotationRepo.Create(ctx, s.db, quotation)
	if err != nil {
		return dto.QuotationResponse{}, err
	}

	return toQuotationResponse(created), nil
}

func (s *quotationService) GetByID(ctx context.Context, id string) (dto.QuotationResponse, error) {
	q, err := s.quotationRepo.GetByID(ctx, s.db, id)
	if err != nil {
		return dto.QuotationResponse{}, dto.ErrQuotationNotFound
	}

	return toQuotationResponse(q), nil
}

func (s *quotationService) Accept(ctx context.Context, quotationID string, clientID string) (dto.QuotationAcceptResponse, error) {
	quotation, err := s.quotationRepo.GetByID(ctx, s.db, quotationID)
	if err != nil {
		return dto.QuotationAcceptResponse{}, dto.ErrQuotationNotFound
	}

	// Validate client owns the project request
	clientUUID, _ := uuid.Parse(clientID)
	if quotation.ProjectRequest.ClientID != clientUUID {
		return dto.QuotationAcceptResponse{}, dto.ErrNotProjectRequestOwner
	}

	if quotation.Status != constants.QUOTATION_STATUS_PENDING {
		return dto.QuotationAcceptResponse{}, dto.ErrQuotationNotPending
	}

	// Atomic: accept quotation + create order + update project request
	tx := s.db.Begin()

	quotation.Status = constants.QUOTATION_STATUS_ACCEPTED
	_, err = s.quotationRepo.Update(ctx, tx, quotation)
	if err != nil {
		tx.Rollback()
		return dto.QuotationAcceptResponse{}, err
	}

	// Update project request status
	pr := quotation.ProjectRequest
	pr.Status = constants.PROJECT_REQUEST_STATUS_IN_PROGRESS
	_, err = s.projectRequestRepo.Update(ctx, tx, pr)
	if err != nil {
		tx.Rollback()
		return dto.QuotationAcceptResponse{}, err
	}

	// Auto-create order
	order := entities.Order{
		ID:            uuid.New(),
		QuotationID:   quotation.ID,
		ClientID:      quotation.ProjectRequest.ClientID,
		DesignerID:    quotation.DesignerID,
		TotalAmount:   quotation.OfferedPrice,
		PaymentStatus: constants.ORDER_PAYMENT_STATUS_UNPAID,
		WorkStatus:    constants.ORDER_WORK_STATUS_ACTIVE,
	}

	createdOrder, err := s.orderRepo.Create(ctx, tx, order)
	if err != nil {
		tx.Rollback()
		return dto.QuotationAcceptResponse{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return dto.QuotationAcceptResponse{}, err
	}

	return dto.QuotationAcceptResponse{
		QuotationID: quotation.ID.String(),
		OrderID:     createdOrder.ID.String(),
		Status:      constants.QUOTATION_STATUS_ACCEPTED,
	}, nil
}

func (s *quotationService) Reject(ctx context.Context, quotationID string, clientID string) error {
	quotation, err := s.quotationRepo.GetByID(ctx, s.db, quotationID)
	if err != nil {
		return dto.ErrQuotationNotFound
	}

	// Validate client owns the project request
	clientUUID, _ := uuid.Parse(clientID)
	if quotation.ProjectRequest.ClientID != clientUUID {
		return dto.ErrNotProjectRequestOwner
	}

	if quotation.Status != constants.QUOTATION_STATUS_PENDING {
		return dto.ErrQuotationNotPending
	}

	quotation.Status = constants.QUOTATION_STATUS_REJECTED
	_, err = s.quotationRepo.Update(ctx, s.db, quotation)
	return err
}

func toQuotationResponse(q entities.Quotation) dto.QuotationResponse {
	return dto.QuotationResponse{
		ID:               q.ID.String(),
		ProjectRequestID: q.ProjectRequestID.String(),
		DesignerID:       q.DesignerID.String(),
		ScopeOfWork:      q.ScopeOfWork,
		OfferedPrice:     q.OfferedPrice.InexactFloat64(),
		DurationDays:     q.DurationDays,
		Status:           q.Status,
	}
}
