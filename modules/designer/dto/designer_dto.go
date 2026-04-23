package dto

import "errors"

const (
	MESSAGE_FAILED_GET_DESIGNER_PROFILE    = "failed get designer profile"
	MESSAGE_FAILED_UPDATE_DESIGNER_PROFILE = "failed update designer profile"
	MESSAGE_SUCCESS_GET_DESIGNER_PROFILE   = "success get designer profile"
	MESSAGE_SUCCESS_UPDATE_DESIGNER_PROFILE = "success update designer profile"
)

var (
	ErrDesignerProfileNotFound = errors.New("designer profile not found")
	ErrNotADesigner            = errors.New("only designer can access this resource")
)

type (
	DesignerProfileUpdateRequest struct {
		Bio               string `json:"bio" binding:"omitempty,max=1000"`
		ExperienceYears   *int   `json:"experience_years" binding:"omitempty,min=0,max=50"`
		IsAvailable       *bool  `json:"is_available" binding:"omitempty"`
		Location          string `json:"location" binding:"omitempty,max=255"`
		BankAccountNumber string `json:"bank_account_number" binding:"omitempty,max=50"`
    }

	DesignerProfileResponse struct {
		ID              string  `json:"id"`
		UserID          string  `json:"user_id"`
		Name            string  `json:"name"`
		ProfilePicture  *string `json:"profile_picture"`
		Bio             string  `json:"bio"`
		ExperienceYears int     `json:"experience_years"`
		IsVerified      bool    `json:"is_verified"`
		IsAvailable     bool    `json:"is_available"`
		Location        string  `json:"location"`
    }
)
