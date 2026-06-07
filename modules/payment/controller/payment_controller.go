package controller

import (
	"errors"
	"net/http"

	"github.com/Rizal-Nurochman/matchnbuild/modules/payment/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/payment/service"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/utils"
	"github.com/gin-gonic/gin"
)

type PaymentController interface {
	CreateSnapToken(ctx *gin.Context)
	GetStatus(ctx *gin.Context)
	Notification(ctx *gin.Context)
}

type paymentController struct {
	service service.PaymentService
}

func NewPaymentController(service service.PaymentService) PaymentController {
	return &paymentController{service: service}
}

func (c *paymentController) CreateSnapToken(ctx *gin.Context) {
	result, err := c.service.CreateSnapToken(
		ctx.Request.Context(),
		ctx.Param("order_id"),
		ctx.MustGet("user_id").(string),
	)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_PAYMENT, err.Error(), nil)
		ctx.JSON(paymentErrorStatus(err), res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_PAYMENT, result)
	ctx.JSON(http.StatusCreated, res)
}

func (c *paymentController) GetStatus(ctx *gin.Context) {
	result, err := c.service.GetStatus(
		ctx.Request.Context(),
		ctx.Param("order_id"),
		ctx.MustGet("user_id").(string),
	)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_PAYMENT, err.Error(), nil)
		ctx.JSON(paymentErrorStatus(err), res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_PAYMENT, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *paymentController) Notification(ctx *gin.Context) {
	var notification dto.NotificationRequest
	if err := ctx.ShouldBindJSON(&notification); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_NOTIFICATION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	if err := c.service.HandleNotification(ctx.Request.Context(), notification); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_NOTIFICATION, err.Error(), nil)
		ctx.JSON(paymentErrorStatus(err), res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_NOTIFICATION, nil)
	ctx.JSON(http.StatusOK, res)
}

func paymentErrorStatus(err error) int {
	switch {
	case errors.Is(err, dto.ErrNotOrderOwner), errors.Is(err, dto.ErrInvalidSignature):
		return http.StatusUnauthorized
	case errors.Is(err, dto.ErrOrderNotFound), errors.Is(err, dto.ErrPaymentNotFound):
		return http.StatusNotFound
	case errors.Is(err, dto.ErrMidtransNotConfigured):
		return http.StatusServiceUnavailable
	default:
		return http.StatusBadRequest
	}
}
