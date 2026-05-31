package controller

import (
	"net/http"

	"github.com/Rizal-Nurochman/matchnbuild/modules/recommendation/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/recommendation/service"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/utils"
	"github.com/gin-gonic/gin"
)

type (
	RecommendationHandler interface {
		GetRecommendations(ctx *gin.Context)
	}

	recommendationHandler struct {
		recommendationService service.RecommendationService
	}
)

func NewRecommendationHandler(svc service.RecommendationService) RecommendationHandler {
	return &recommendationHandler{recommendationService: svc}
}

func (h *recommendationHandler) GetRecommendations(ctx *gin.Context) {
	var req dto.RecommendationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_RECOMMENDATION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	results, err := h.recommendationService.GetRecommendations(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_RECOMMENDATION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_RECOMMENDATION, results)
	ctx.JSON(http.StatusOK, res)
}