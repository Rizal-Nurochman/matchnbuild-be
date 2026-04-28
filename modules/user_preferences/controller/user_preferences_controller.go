package controller

import (
	"net/http"

	"github.com/Rizal-Nurochman/matchnbuild/modules/user_preferences/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/user_preferences/service"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type(
	userPreferenceHandler struct {
		userPreferenceService service.UserPreferenceService
		db 										*gorm.DB
	}

	UserPreferenceHandler interface {
		Create(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		Update(ctx *gin.Context)
	}
)

func NewUserPreferenceHandler(injector *do.Injector, userPrefenceService service.UserPreferenceService) UserPreferenceHandler {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &userPreferenceHandler{userPreferenceService: userPrefenceService, db: db}
}

func (h *userPreferenceHandler) Create(ctx *gin.Context) {
	var req dto.CreatePreferenceRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_PREFERENCE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	userID := ctx.MustGet("user_id").(string)
	result, err := h.userPreferenceService.Create(ctx.Request.Context(), userID, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.ErrPreferenceNotFound.Error(), err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_PREFERENCE, result)
	ctx.JSON(http.StatusOK, res)
}

func (h *userPreferenceHandler) GetByID(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)

	preference, err := h.userPreferenceService.GetByUserID(ctx.Request.Context(), userID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_PREFERENCE, err.Error(), nil)
    ctx.JSON(http.StatusBadRequest, res)
    return
  }

  res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_PREFERENCE, preference)
  ctx.JSON(http.StatusOK, res)
}

func (h *userPreferenceHandler) Update(ctx *gin.Context)  {
	var req dto.UpdatePreferenceRequest
	userID := ctx.MustGet("user_id").(string)
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_PREFERENCE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.userPreferenceService.Update(ctx.Request.Context(), userID, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_PREFERENCE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_FAILED_UPDATE_PREFERENCE, result)
	ctx.JSON(http.StatusOK, res)
}