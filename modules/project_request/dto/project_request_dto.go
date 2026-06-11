package dto

import "errors"

const (
	MESSAGE_FAILED_CREATE_PROJECT_REQUEST = "failed create project request"
	MESSAGE_FAILED_GET_PROJECT_REQUEST    = "failed get project request"
	MESSAGE_FAILED_UPDATE_PROJECT_REQUEST = "failed update project request"
	MESSAGE_FAILED_DELETE_PROJECT_REQUEST = "failed delete project request"
	MESSAGE_SUCCESS_CREATE_PROJECT_REQUEST = "success create project request"
	MESSAGE_SUCCESS_GET_PROJECT_REQUEST    = "success get project request"
	MESSAGE_SUCCESS_UPDATE_PROJECT_REQUEST = "success update project request"
	MESSAGE_SUCCESS_DELETE_PROJECT_REQUEST = "success delete project request"
)

var (
	ErrProjectRequestNotFound    = errors.New("project request not found")
	ErrDesignerProfileNotFound   = errors.New("designer profile not found")
	ErrCannotRequestOwnDesign    = errors.New("cannot create project request to yourself")
)

type (
	ProjectRequestCreateRequest struct {
		DesignerID       string  `json:"designer_id" binding:"required"`
		Description      string  `json:"description" binding:"required"`
		InitialBudget    float64 `json:"initial_budget" binding:"required,min=0"`
		AreaSize         float64 `json:"area_size" binding:"required,min=0"`
		LocationPhotoURL string  `json:"location_photo_url"`
		LayoutSketchURL  string  `json:"layout_sketch_url"`
	}

	ClientInfo struct {
		ClientID string `json:"client_id"`
		Name     string `json:"name"`
	}

	DesignerInfo struct {
		DesignerID string `json:"designer_id"`
		Name       string `json:"name"`
	}

	OrderInfo struct {
		ID            string `json:"id"`
		PaymentStatus string `json:"payment_status"`
		WorkStatus    string `json:"work_status"`
	}

	QuotationInfo struct {
		ID           string     `json:"id"`
		ScopeOfWork  string     `json:"scope_of_work"`
		OfferedPrice float64    `json:"offered_price"`
		DurationDays int        `json:"duration_days"`
		Status       string     `json:"status"`
		Order        *OrderInfo `json:"order,omitempty"`
	}

	ProjectRequestResponse struct {
		ID               string          `json:"id"`
		Client           ClientInfo      `json:"client"`
		Designer         DesignerInfo    `json:"designer"`
		Description      string          `json:"description"`
		InitialBudget    float64         `json:"initial_budget"`
		AreaSize         float64         `json:"area_size"`
		LocationPhotoURL string          `json:"location_photo_url"`
		LayoutSketchURL  string          `json:"layout_sketch_url"`
		Status           string          `json:"status"`
		Quotations       []QuotationInfo `json:"quotations"`
		ConversationID   string          `json:"conversation_id,omitempty"`
	}
)
