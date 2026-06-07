package dto

import "errors"

const (
	MESSAGE_FAILED_CREATE_PAYMENT = "failed to create payment"
	MESSAGE_FAILED_GET_PAYMENT    = "failed to get payment status"
	MESSAGE_FAILED_NOTIFICATION   = "failed to process payment notification"

	MESSAGE_SUCCESS_CREATE_PAYMENT = "payment transaction created"
	MESSAGE_SUCCESS_GET_PAYMENT    = "payment status retrieved"
	MESSAGE_SUCCESS_NOTIFICATION   = "payment notification processed"
)

var (
	ErrMidtransNotConfigured = errors.New("midtrans is not configured")
	ErrOrderNotFound         = errors.New("order not found")
	ErrPaymentNotFound       = errors.New("payment not found")
	ErrNotOrderOwner         = errors.New("you are not the owner of this order")
	ErrOrderAlreadyPaid      = errors.New("order is already paid")
	ErrInvalidAmount         = errors.New("order amount must be a positive whole IDR amount")
	ErrInvalidSignature      = errors.New("invalid Midtrans notification signature")
	ErrAmountMismatch        = errors.New("payment amount does not match the order amount")
	ErrInvalidNotification   = errors.New("invalid Midtrans notification")
)

type SnapTokenResponse struct {
	PaymentID       string `json:"payment_id"`
	OrderID         string `json:"order_id"`
	MidtransOrderID string `json:"midtrans_order_id"`
	SnapToken       string `json:"snap_token"`
	RedirectURL     string `json:"redirect_url"`
	ClientKey       string `json:"client_key"`
	Environment     string `json:"environment"`
	ExpiresAt       string `json:"expires_at"`
}

type PaymentStatusResponse struct {
	PaymentID       string `json:"payment_id"`
	OrderID         string `json:"order_id"`
	MidtransOrderID string `json:"midtrans_order_id"`
	Amount          string `json:"amount"`
	PaymentType     string `json:"payment_type"`
	Status          string `json:"status"`
	TransactionID   string `json:"transaction_id"`
	PaidAt          string `json:"paid_at,omitempty"`
	ExpiredAt       string `json:"expired_at,omitempty"`
	OrderStatus     string `json:"order_payment_status"`
}

type NotificationRequest struct {
	OrderID           string `json:"order_id" binding:"required"`
	StatusCode        string `json:"status_code" binding:"required"`
	GrossAmount       string `json:"gross_amount" binding:"required"`
	SignatureKey      string `json:"signature_key" binding:"required"`
	TransactionID     string `json:"transaction_id"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
}
