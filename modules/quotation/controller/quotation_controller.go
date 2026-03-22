package controller

import (
	"net/http"

	"github.com/Rizal-Nurochman/matchnbuild/modules/quotation/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/quotation/service"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/utils"
	"github.com/gin-gonic/gin"
)

type (
	QuotationController interface {
		Create(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		Accept(ctx *gin.Context)
		Reject(ctx *gin.Context)
	}

	quotationController struct {
		quotationService service.QuotationService
	}
)

func NewQuotationController(qs service.QuotationService) QuotationController {
	return &quotationController{
		quotationService: qs,
	}
}

func (c *quotationController) Create(ctx *gin.Context) {
	var req dto.QuotationCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed("failed to get data from body", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	designerID := ctx.MustGet("user_id").(string)

	result, err := c.quotationService.Create(ctx.Request.Context(), req, designerID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_QUOTATION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_QUOTATION, result)
	ctx.JSON(http.StatusCreated, res)
}

func (c *quotationController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")

	result, err := c.quotationService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_QUOTATION, err.Error(), nil)
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_QUOTATION, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *quotationController) Accept(ctx *gin.Context) {
	quotationID := ctx.Param("id")
	clientID := ctx.MustGet("user_id").(string)

	result, err := c.quotationService.Accept(ctx.Request.Context(), quotationID, clientID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_ACCEPT_QUOTATION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_ACCEPT_QUOTATION, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *quotationController) Reject(ctx *gin.Context) {
	quotationID := ctx.Param("id")
	clientID := ctx.MustGet("user_id").(string)

	err := c.quotationService.Reject(ctx.Request.Context(), quotationID, clientID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_REJECT_QUOTATION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_REJECT_QUOTATION, nil)
	ctx.JSON(http.StatusOK, res)
}
