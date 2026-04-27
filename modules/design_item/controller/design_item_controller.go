package controller

import (
	"net/http"

	"github.com/Rizal-Nurochman/matchnbuild/modules/design_item/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/design_item/service"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/utils"
	"github.com/gin-gonic/gin"
)

type (
	designItemHandler struct {
		designItemService service.DesignItemService
	}

	DesignItemHandler interface {
		Create(ctx *gin.Context)
		GetAll(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		GetMyItems(ctx *gin.Context)
		Update(ctx *gin.Context)
		Delete(ctx *gin.Context)
	}
)

func NewDesignItemHandler(designItemService service.DesignItemService) DesignItemHandler {
	return &designItemHandler{designItemService: designItemService}
}

func (h *designItemHandler) Create(ctx *gin.Context) {
	var req dto.DesignItemCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_DESIGN_ITEM, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	userID := ctx.MustGet("user_id").(string)
	result, err := h.designItemService.Create(ctx.Request.Context(), userID, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_DESIGN_ITEM, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_DESIGN_ITEM, result)
	ctx.JSON(http.StatusCreated, res)
}

func (h *designItemHandler) GetAll(ctx *gin.Context) {
	items, err := h.designItemService.GetAll(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DESIGN_ITEM, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_DESIGN_ITEM, items)
	ctx.JSON(http.StatusOK, res)
}

func (h *designItemHandler) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")

	item, err := h.designItemService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DESIGN_ITEM, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_DESIGN_ITEM, item)
	ctx.JSON(http.StatusOK, res)
}

func (h *designItemHandler) GetMyItems(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)

	items, err := h.designItemService.GetMyItems(ctx.Request.Context(), userID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DESIGN_ITEM, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_DESIGN_ITEM, items)
	ctx.JSON(http.StatusOK, res)
}

func (h *designItemHandler) Update(ctx *gin.Context) {
	var req dto.DesignItemUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_DESIGN_ITEM, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	userID := ctx.MustGet("user_id").(string)
	designItemID := ctx.Param("id")

	result, err := h.designItemService.Update(ctx.Request.Context(), userID, designItemID, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_DESIGN_ITEM, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_DESIGN_ITEM, result)
	ctx.JSON(http.StatusOK, res)
}

func (h *designItemHandler) Delete(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	designItemID := ctx.Param("id")

	err := h.designItemService.Delete(ctx.Request.Context(), userID, designItemID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_DESIGN_ITEM, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_DESIGN_ITEM, nil)
	ctx.JSON(http.StatusOK, res)
}