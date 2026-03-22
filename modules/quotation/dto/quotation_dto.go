package dto

import "errors"

const (
	MESSAGE_FAILED_CREATE_QUOTATION  = "failed create quotation"
	MESSAGE_FAILED_GET_QUOTATION     = "failed get quotation"
	MESSAGE_FAILED_ACCEPT_QUOTATION  = "failed accept quotation"
	MESSAGE_FAILED_REJECT_QUOTATION  = "failed reject quotation"
	MESSAGE_SUCCESS_CREATE_QUOTATION = "success create quotation"
	MESSAGE_SUCCESS_GET_QUOTATION    = "success get quotation"
	MESSAGE_SUCCESS_ACCEPT_QUOTATION = "success accept quotation"
	MESSAGE_SUCCESS_REJECT_QUOTATION = "success reject quotation"
)

var (
	ErrQuotationNotFound         = errors.New("quotation not found")
	ErrQuotationAlreadyExists    = errors.New("quotation already exists for this project request")
	ErrQuotationNotPending       = errors.New("quotation is not in pending status")
	ErrNotProjectRequestOwner    = errors.New("you are not the owner of this project request")
	ErrNotDesignerOfRequest      = errors.New("you are not the designer of this project request")
	ErrProjectRequestNotOpen     = errors.New("project request is not open")
)

type (
	QuotationCreateRequest struct {
		ProjectRequestID string  `json:"project_request_id" binding:"required"`
		ScopeOfWork      string  `json:"scope_of_work" binding:"required"`
		OfferedPrice     float64 `json:"offered_price" binding:"required,min=0"`
		DurationDays     int     `json:"duration_days" binding:"required,min=1"`
	}

	QuotationResponse struct {
		ID               string  `json:"id"`
		ProjectRequestID string  `json:"project_request_id"`
		DesignerID       string  `json:"designer_id"`
		ScopeOfWork      string  `json:"scope_of_work"`
		OfferedPrice     float64 `json:"offered_price"`
		DurationDays     int     `json:"duration_days"`
		Status           string  `json:"status"`
	}

	QuotationAcceptResponse struct {
		QuotationID string `json:"quotation_id"`
		OrderID     string `json:"order_id"`
		Status      string `json:"status"`
	}
)
