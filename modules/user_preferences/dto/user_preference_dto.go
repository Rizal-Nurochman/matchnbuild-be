package dto

import (
	"errors"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
)

const (
	// Failed
	MESSAGE_FAILED_CREATE_PREFERENCE = "failed create preference"
	MESSAGE_FAILED_GET_PREFERENCE    = "failed get preference"
	MESSAGE_FAILED_UPDATE_PREFERENCE = "failed update preference"

	// Success
	MESSAGE_SUCCESS_CREATE_PREFERENCE = "success create preference"
	MESSAGE_SUCCESS_GET_PREFERENCE    = "success get preference"
	MESSAGE_SUCCESS_UPDATE_PREFERENCE = "success update preference"
)

var (
	ErrPreferenceNotFound      = errors.New("preference not found")
	ErrPreferenceAlreadyExist  = errors.New("preference already exists")
)

type (
	CreatePreferenceRequest struct {
		PreferredStyle    string  `json:"preferred_style" binding:"required"`
		BudgetMin         float64 `json:"budget_min" binding:"required,gt=0"`
		BudgetMax         float64 `json:"budget_max" binding:"required,gt=0"`
		PreferredLocation string  `json:"preferred_location" binding:"required"`
	}

	UpdatePreferenceRequest struct {
		PreferredStyle    *string   `json:"preferred_style" binding:"omitempty"`
		BudgetMin         *float64 `json:"budget_min" binding:"omitempty,gt=0"`
		BudgetMax         *float64 `json:"budget_max" binding:"omitempty,gt=0"`
		PreferredLocation *string   `json:"preferred_location" binding:"omitempty"`
	}

	PreferenceResponse struct {
		ID                string `json:"id"`
		PreferredStyle    string `json:"preferred_style"`
		BudgetMin         string `json:"budget_min"`
		BudgetMax         string `json:"budget_max"`
		PreferredLocation string `json:"preferred_location"`
		IsOnboarded       bool   `json:"is_onboarded"`
	}
)

func ToUserPreferenceResponse(preference entities.UserPreference) PreferenceResponse {
	return PreferenceResponse{
		ID: preference.ID.String(),
		PreferredStyle: preference.PreferredStyle,
		BudgetMin: preference.BudgetMin.String(),
		BudgetMax: preference.BudgetMax.String(),
		PreferredLocation: preference.PreferredLocation,
		IsOnboarded: preference.IsOnboarded,
	}
}