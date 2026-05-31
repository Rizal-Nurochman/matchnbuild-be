package dto

import "github.com/Rizal-Nurochman/matchnbuild/modules/design_item/dto"

const (
	MESSAGE_FAILED_GET_RECOMMENDATION = "failed get recommendation"
	MESSAGE_SUCCESS_GET_RECOMMENDATION = "success get recommendation"
)

type (
	RecommendationRequest struct {
		Category        string   `json:"category" binding:"required,oneof=architecture interior"`
		Style           string   `json:"style" binding:"required"`
		Location        string   `json:"location" binding:"required"`
		EstimatedBudget float64  `json:"estimated_budget" binding:"required,gt=0"`
		LandAreaMax     *float64 `json:"land_area_max" binding:"omitempty,gt=0"`
		LandAreaMin     *float64 `json:"land_area_min" binding:"omitempty,gt=0"`
		BuildingArea    *float64 `json:"building_area" binding:"omitempty,gt=0"`
		NumFloors       *int     `json:"num_floors" binding:"omitempty,min=1,max=10"`
		NumBedrooms     *int     `json:"num_bedrooms" binding:"omitempty,min=0,max=20"`
		RoomType        *string  `json:"room_type" binding:"omitempty,max=50"`
		RoomArea        *float64 `json:"room_area" binding:"omitempty,gt=0"`
		Limit           *int     `json:"limit" binding:"omitempty"`
	}

	MLRecommendationResponse struct {
		ItemID string  `json:"item_id"`
		Score  float64 `json:"score"`
	}

	RecommendationResponse struct {
		Score float64 `json:"score"`
		Item  dto.DesignItemResponse	`json:"item"`
	}
)