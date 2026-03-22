package controller

import (
	"net/http"

	"github.com/Rizal-Nurochman/matchnbuild/modules/project_request/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/project_request/service"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/utils"
	"github.com/gin-gonic/gin"
)

type (
	ProjectRequestController interface {
		Create(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		GetMyRequests(ctx *gin.Context)
		GetMyIncomingRequests(ctx *gin.Context)
	}

	projectRequestController struct {
		projectRequestService service.ProjectRequestService
	}
)

func NewProjectRequestController(prs service.ProjectRequestService) ProjectRequestController {
	return &projectRequestController{
		projectRequestService: prs,
	}
}

func (c *projectRequestController) Create(ctx *gin.Context) {
	var req dto.ProjectRequestCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed("failed to get data from body", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	clientID := ctx.MustGet("user_id").(string)

	result, err := c.projectRequestService.Create(ctx.Request.Context(), req, clientID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_PROJECT_REQUEST, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_PROJECT_REQUEST, result)
	ctx.JSON(http.StatusCreated, res)
}

func (c *projectRequestController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")

	result, err := c.projectRequestService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_PROJECT_REQUEST, err.Error(), nil)
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_PROJECT_REQUEST, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *projectRequestController) GetMyRequests(ctx *gin.Context) {
	clientID := ctx.MustGet("user_id").(string)

	result, err := c.projectRequestService.GetByClientID(ctx.Request.Context(), clientID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_PROJECT_REQUEST, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_PROJECT_REQUEST, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *projectRequestController) GetMyIncomingRequests(ctx *gin.Context) {
	designerID := ctx.MustGet("user_id").(string)

	result, err := c.projectRequestService.GetByDesignerID(ctx.Request.Context(), designerID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_PROJECT_REQUEST, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_PROJECT_REQUEST, result)
	ctx.JSON(http.StatusOK, res)
}
