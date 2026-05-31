package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Rizal-Nurochman/matchnbuild/modules/design_item/dto"
	designItemRepo "github.com/Rizal-Nurochman/matchnbuild/modules/design_item/repository"
	recDto "github.com/Rizal-Nurochman/matchnbuild/modules/recommendation/dto"
)

type (
	RecommendationService interface {
		GetRecommendations(ctx context.Context, req recDto.RecommendationRequest) ([]recDto.RecommendationResponse, error)
	}

	recommendationService struct {
		designItemRepo designItemRepo.DesignItemRepository
		httpClient     *http.Client
	}
)

func NewRecommendationService(designItemRepo designItemRepo.DesignItemRepository) RecommendationService {
	return &recommendationService{
		designItemRepo: designItemRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *recommendationService) GetRecommendations(ctx context.Context, req recDto.RecommendationRequest) ([]recDto.RecommendationResponse, error) {
	if err := validateByCategory(req); err != nil {
		return nil, err
	}
	mlResults, err := s.callMLService(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ml service error: %w", err)
	}
	var results []recDto.RecommendationResponse
	for _, ml := range mlResults {
		item, err := s.designItemRepo.GetByID(ctx, nil, ml.ItemID)
		if err != nil {
			continue
		}

		results = append(results, recDto.RecommendationResponse{
			Score: ml.Score,
			Item:  dto.ToDesignItemResponse(item),
		})
	}

	if results == nil {
		results = []recDto.RecommendationResponse{}
	}

	return results, nil
}

func validateByCategory(req recDto.RecommendationRequest) error {
	switch req.Category {
	case "architecture":
		if req.LandAreaMin == nil || req.BuildingArea == nil || req.NumFloors == nil {
			return errors.New("land_area_min, building_area, dan num_floors wajib diisi untuk arsitektur")
		}
	case "interior":
		if req.RoomType == nil || req.RoomArea == nil {
			return errors.New("room_type dan room_area wajib diisi untuk interior")
		}
	}
	return nil
}

func (s *recommendationService) callMLService(ctx context.Context, req recDto.RecommendationRequest) ([]recDto.MLRecommendationResponse, error) {
	mlURL := os.Getenv("ML_SERVICE_URL")
	if mlURL == "" {
		mlURL = "http://localhost:8000"
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, mlURL+"/api/recommend", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ml service returned status %d", resp.StatusCode)
	}

	var results []recDto.MLRecommendationResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	return results, nil
}