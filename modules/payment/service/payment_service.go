package service

import (
	"context"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Rizal-Nurochman/matchnbuild/config"
	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"github.com/Rizal-Nurochman/matchnbuild/modules/payment/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/payment/gateway"
	"github.com/Rizal-Nurochman/matchnbuild/modules/payment/repository"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PaymentService interface {
	CreateSnapToken(ctx context.Context, orderID string, userID string) (dto.SnapTokenResponse, error)
	GetStatus(ctx context.Context, orderID string, userID string) (dto.PaymentStatusResponse, error)
	HandleNotification(ctx context.Context, notification dto.NotificationRequest) error
}

type paymentService struct {
	repository repository.PaymentRepository
	gateway    gateway.MidtransGateway
	config     config.MidtransConfig
	db         *gorm.DB
}

func NewPaymentService(
	repository repository.PaymentRepository,
	gateway gateway.MidtransGateway,
	cfg config.MidtransConfig,
	db *gorm.DB,
) PaymentService {
	return &paymentService{
		repository: repository,
		gateway:    gateway,
		config:     cfg,
		db:         db,
	}
}

func (s *paymentService) CreateSnapToken(ctx context.Context, orderID string, userID string) (dto.SnapTokenResponse, error) {
	if s.config.ServerKey == "" || s.config.ClientKey == "" {
		return dto.SnapTokenResponse{}, dto.ErrMidtransNotConfigured
	}

	var result dto.SnapTokenResponse
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := s.repository.GetOrderByID(ctx, tx, orderID, true)
		if err != nil {
			return dto.ErrOrderNotFound
		}
		if order.ClientID.String() != userID {
			return dto.ErrNotOrderOwner
		}
		if order.PaymentStatus == constants.ORDER_PAYMENT_STATUS_PAID {
			return dto.ErrOrderAlreadyPaid
		}

		amount, err := amountToIDR(order.TotalAmount)
		if err != nil {
			return err
		}

		existing, err := s.repository.GetLatestByOrderID(ctx, tx, orderID)
		if err == nil && existing.Status == constants.PAYMENT_STATUS_PENDING &&
			existing.SnapToken != "" && existing.ExpiredAt != nil && existing.ExpiredAt.After(time.Now()) {
			result = s.toSnapTokenResponse(existing, "")
			return nil
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		now := time.Now()
		expiresAt := now.Add(time.Duration(s.config.ExpiryMinutes) * time.Minute)
		midtransOrderID := buildMidtransOrderID(order.ID, now)
		payment := entities.Payment{
			ID:              uuid.New(),
			OrderID:         order.ID,
			MidtransOrderID: midtransOrderID,
			Amount:          order.TotalAmount,
			Status:          constants.PAYMENT_STATUS_PENDING,
			ExpiredAt:       &expiresAt,
		}

		payment, err = s.repository.Create(ctx, tx, payment)
		if err != nil {
			return err
		}

		request := &snap.Request{
			TransactionDetails: midtrans.TransactionDetails{
				OrderID:  midtransOrderID,
				GrossAmt: amount,
			},
			Items: &[]midtrans.ItemDetails{{
				ID:    order.ID.String(),
				Name:  "Match & Build Design Service",
				Price: amount,
				Qty:   1,
			}},
			CustomerDetail: &midtrans.CustomerDetails{
				FName: order.Client.Name,
				Email: order.Client.Email,
			},
			CreditCard: &snap.CreditCardDetails{Secure: true},
			Expiry: &snap.ExpiryDetails{
				StartTime: now.Format("2006-01-02 15:04:05 -0700"),
				Unit:      "minute",
				Duration:  s.config.ExpiryMinutes,
			},
		}
		if s.config.FinishURL != "" {
			request.Callbacks = &snap.Callbacks{Finish: s.config.FinishURL}
		}

		snapResponse, gatewayErr := s.gateway.CreateSnapTransaction(ctx, request)
		if gatewayErr != nil {
			return gatewayErr
		}

		payment.SnapToken = snapResponse.Token
		if err := s.repository.Update(ctx, tx, payment); err != nil {
			return err
		}

		result = s.toSnapTokenResponse(payment, snapResponse.RedirectURL)
		return nil
	})

	return result, err
}

func (s *paymentService) GetStatus(ctx context.Context, orderID string, userID string) (dto.PaymentStatusResponse, error) {
	order, err := s.repository.GetOrderByID(ctx, s.db, orderID, false)
	if err != nil {
		return dto.PaymentStatusResponse{}, dto.ErrOrderNotFound
	}
	if order.ClientID.String() != userID {
		return dto.PaymentStatusResponse{}, dto.ErrNotOrderOwner
	}

	payment, err := s.repository.GetLatestByOrderID(ctx, s.db, orderID)
	if err != nil {
		return dto.PaymentStatusResponse{}, dto.ErrPaymentNotFound
	}

	if payment.MidtransOrderID != "" {
		status, gatewayErr := s.gateway.CheckTransaction(ctx, payment.MidtransOrderID)
		if gatewayErr == nil {
			if err := s.applyStatus(ctx, status); err != nil {
				return dto.PaymentStatusResponse{}, err
			}
			payment, _ = s.repository.GetByMidtransOrderID(ctx, s.db, payment.MidtransOrderID, false)
			order, _ = s.repository.GetOrderByID(ctx, s.db, orderID, false)
		}
	}

	return toPaymentStatusResponse(payment, order.PaymentStatus), nil
}

func (s *paymentService) HandleNotification(ctx context.Context, notification dto.NotificationRequest) error {
	if s.config.ServerKey == "" {
		return dto.ErrMidtransNotConfigured
	}
	if !verifySignature(notification, s.config.ServerKey) {
		return dto.ErrInvalidSignature
	}

	status, err := s.gateway.CheckTransaction(ctx, notification.OrderID)
	if err != nil {
		return err
	}
	if status.OrderID != notification.OrderID {
		return dto.ErrInvalidNotification
	}

	return s.applyStatus(ctx, status)
}

func (s *paymentService) applyStatus(ctx context.Context, status *coreapi.TransactionStatusResponse) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		payment, err := s.repository.GetByMidtransOrderID(ctx, tx, status.OrderID, true)
		if err != nil {
			return dto.ErrPaymentNotFound
		}

		grossAmount, err := decimal.NewFromString(status.GrossAmount)
		if err != nil || !grossAmount.Equal(payment.Amount) {
			return dto.ErrAmountMismatch
		}

		order, err := s.repository.GetOrderByID(ctx, tx, payment.OrderID.String(), true)
		if err != nil {
			return dto.ErrOrderNotFound
		}

		payment.PaymentType = status.PaymentType
		payment.PaymentMethod = status.PaymentType
		payment.MidtransTransID = status.TransactionID
		payment.Status = normalizePaymentStatus(status.TransactionStatus)

		if isPaid(status.TransactionStatus, status.FraudStatus) {
			order.PaymentStatus = constants.ORDER_PAYMENT_STATUS_PAID
			paidAt := parseMidtransTime(status.SettlementTime)
			if paidAt == nil {
				now := time.Now()
				paidAt = &now
			}
			payment.PaidAt = paidAt
		} else if status.TransactionStatus == "refund" {
			order.PaymentStatus = constants.ORDER_PAYMENT_STATUS_REFUNDED
		}

		if expiry := parseMidtransTime(status.ExpiryTime); expiry != nil {
			payment.ExpiredAt = expiry
		}

		if err := s.repository.Update(ctx, tx, payment); err != nil {
			return err
		}
		return s.repository.UpdateOrder(ctx, tx, order)
	})
}

func (s *paymentService) toSnapTokenResponse(payment entities.Payment, redirectURL string) dto.SnapTokenResponse {
	environment := "sandbox"
	if s.config.Environment == midtrans.Production {
		environment = "production"
	}

	expiresAt := ""
	if payment.ExpiredAt != nil {
		expiresAt = payment.ExpiredAt.Format(time.RFC3339)
	}

	return dto.SnapTokenResponse{
		PaymentID:       payment.ID.String(),
		OrderID:         payment.OrderID.String(),
		MidtransOrderID: payment.MidtransOrderID,
		SnapToken:       payment.SnapToken,
		RedirectURL:     redirectURL,
		ClientKey:       s.config.ClientKey,
		Environment:     environment,
		ExpiresAt:       expiresAt,
	}
}

func amountToIDR(amount decimal.Decimal) (int64, error) {
	if amount.LessThanOrEqual(decimal.Zero) || !amount.Equal(amount.Truncate(0)) {
		return 0, dto.ErrInvalidAmount
	}
	return amount.IntPart(), nil
}

func buildMidtransOrderID(orderID uuid.UUID, now time.Time) string {
	uuidPart := strings.ReplaceAll(orderID.String(), "-", "")[:8]
	timePart := strings.ToUpper(strconv.FormatInt(now.Unix(), 36))
	return fmt.Sprintf("MNB-%s-%s", timePart, strings.ToUpper(uuidPart))
}

func verifySignature(notification dto.NotificationRequest, serverKey string) bool {
	raw := notification.OrderID + notification.StatusCode + notification.GrossAmount + serverKey
	hash := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(hash[:])
	actual := strings.ToLower(notification.SignatureKey)
	return len(expected) == len(actual) &&
		subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}

func isPaid(transactionStatus string, fraudStatus string) bool {
	return transactionStatus == "settlement" ||
		(transactionStatus == "capture" && (fraudStatus == "" || fraudStatus == "accept"))
}

func normalizePaymentStatus(status string) string {
	switch status {
	case "capture":
		return constants.PAYMENT_STATUS_CAPTURE
	case "settlement":
		return constants.PAYMENT_STATUS_SETTLEMENT
	case "expire":
		return constants.PAYMENT_STATUS_EXPIRE
	case "cancel":
		return constants.PAYMENT_STATUS_CANCEL
	case "deny":
		return constants.PAYMENT_STATUS_DENY
	case "refund":
		return constants.PAYMENT_STATUS_REFUND
	default:
		return constants.PAYMENT_STATUS_PENDING
	}
}

func parseMidtransTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.FixedZone("WIB", 7*60*60))
	if err != nil {
		return nil
	}
	return &parsed
}

func toPaymentStatusResponse(payment entities.Payment, orderStatus string) dto.PaymentStatusResponse {
	response := dto.PaymentStatusResponse{
		PaymentID:       payment.ID.String(),
		OrderID:         payment.OrderID.String(),
		MidtransOrderID: payment.MidtransOrderID,
		Amount:          payment.Amount.StringFixed(2),
		PaymentType:     payment.PaymentType,
		Status:          payment.Status,
		TransactionID:   payment.MidtransTransID,
		OrderStatus:     orderStatus,
	}
	if payment.PaidAt != nil {
		response.PaidAt = payment.PaidAt.Format(time.RFC3339)
	}
	if payment.ExpiredAt != nil {
		response.ExpiredAt = payment.ExpiredAt.Format(time.RFC3339)
	}
	return response
}
