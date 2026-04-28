package dto

import (
	"errors"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
)

const (
	MESSAGE_FAILED_CREATE_DESIGN_ITEM = "failed create design item"
	MESSAGE_FAILED_GET_DESIGN_ITEM    = "failed get design item"
	MESSAGE_FAILED_UPDATE_DESIGN_ITEM = "failed update design item"
	MESSAGE_FAILED_DELETE_DESIGN_ITEM = "failed delete design item"

	MESSAGE_SUCCESS_CREATE_DESIGN_ITEM = "success create design item"
	MESSAGE_SUCCESS_GET_DESIGN_ITEM    = "success get design item"
	MESSAGE_SUCCESS_UPDATE_DESIGN_ITEM = "success update design item"
	MESSAGE_SUCCESS_DELETE_DESIGN_ITEM = "success delete design item"

	MESSAGE_FAILED_GET_FEATURES = "failed get features"
	MESSGAE_SUCCESS_GET_FEATURES = "success get features"
)

var (
	ErrDesignItemNotFound = errors.New("design item not found")
	ErrNotDesignItemOwner = errors.New("you are not the owner of this design item")
)

type (
	DesignItemCreateRequest struct {
		Title           string   `json:"title" binding:"required,max=200"`
		Description     string   `json:"description" binding:"omitempty,max=2000"`
		Style           string   `json:"style" binding:"required,max=50"`
		Category 				string			`json:"category" binding:"required,max=50"`
		LandAreaMin     *float64  `json:"land_area_min" binding:"omitempty,gt=0"`
		LandAreaMax     *float64    `json:"land_area_max" binding:"omitempty,gt=0"`
		BuildingArea    *float64  `json:"building_area" binding:"omitempty,gt=0"`
		NumFloors       *int      `json:"num_floors" binding:"omitempty,min=1,max=10"`
		NumBedrooms     *int     `json:"num_bedrooms" binding:"omitempty,min=0,max=20"`
		RoomType        *string    `json:"room_type" binding:"omitempty,max=50"`
		RoomArea			  *float64    `json:"room_area" binding:"omitempty,gt=0"`
		EstimatedBudget float64  `json:"estimated_budget" binding:"required,gt=0"`
		PriceStartFrom  float64  `json:"price_start_from" binding:"required,gt=0"`
		ImageURL        string   `json:"image_url" binding:"omitempty,url"`
		FeatureIDs      []string `json:"feature_ids" binding:"omitempty"`
	}

	DesignItemUpdateRequest struct {
		Title           string   `json:"title" binding:"required,max=200"`
		Description     string   `json:"description" binding:"omitempty,max=2000"`
		Style           string   `json:"style" binding:"required,max=50"`
		Category 				string			`json:"category" binding:"required,max=50"`
		LandAreaMin     *float64  `json:"land_area_min" binding:"omitempty,gt=0"`
		LandAreaMax     *float64    `json:"land_area_max" binding:"omitempty,gt=0"`
		BuildingArea    *float64  `json:"building_area" binding:"omitempty,gt=0"`
		NumFloors       *int      `json:"num_floors" binding:"omitempty,min=1,max=10"`
		NumBedrooms     *int     `json:"num_bedrooms" binding:"omitempty,min=0,max=20"`
		RoomType        *string    `json:"room_type" binding:"omitempty,max=50"`
		RoomArea			  *float64    `json:"room_area" binding:"omitempty,gt=0"`
		EstimatedBudget float64  `json:"estimated_budget" binding:"required,gt=0"`
		PriceStartFrom  float64  `json:"price_start_from" binding:"required,gt=0"`
		ImageURL        string   `json:"image_url" binding:"omitempty,url"`
		FeatureIDs      []string `json:"feature_ids" binding:"omitempty"`
	}

	FeatureResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	DesignItemResponse struct {
		ID              string            `json:"id"`
		DesignerID      string            `json:"designer_id"`
		DesignerName    string            `json:"designer_name"`
		Title           string            `json:"title"`
		Description     string            `json:"description"`
		Style           string            `json:"style"`
		Category				string						`json:"category"`
		LandAreaMin     *float64           `json:"land_area_min"`
		LandAreaMax			*float64           `json:"land_area_max"`
		BuildingArea    *float64           `json:"building_area"`
		NumFloors       *int               `json:"num_floors"`
		NumBedrooms     *int              `json:"num_bedrooms"`
		RoomType 				*string						`json:"room_type"`
		RoomArea				*float64	          `json:"room_area"`		
		EstimatedBudget string            `json:"estimated_budget"`
		PriceStartFrom  string            `json:"price_start_from"`
		ImageURL        string            `json:"image_url"`
		Features        []FeatureResponse `json:"features"`
	}
)

func ToDesignItemResponse(item entities.DesignItem) DesignItemResponse {
	var features []FeatureResponse
	for _, f := range item.Features {
		features = append(features, FeatureResponse{
			ID:   f.FeatureID.String(),
			Name: f.Feature.Name,
		})
	}

	return DesignItemResponse{
		ID:              item.ID.String(),
		DesignerID:      item.DesignerID.String(),
		DesignerName:    item.Designer.User.Name,
		Title:           item.Title,
		Description:     item.Description,
		Style:           item.Style,
		Category:        item.Category,
		LandAreaMin:     item.LandAreaMin,
		LandAreaMax:     item.LandAreaMax,
		BuildingArea:    item.BuildingArea,
		NumFloors:       item.NumFloors,
		NumBedrooms:     item.NumBedrooms,
		RoomType:        item.RoomType,
		RoomArea:        item.RoomArea,
		EstimatedBudget: item.EstimatedBudget.String(),
		PriceStartFrom:  item.PriceStartFrom.String(),
		ImageURL:        item.ImageURL,
		Features:        features,
	}
}