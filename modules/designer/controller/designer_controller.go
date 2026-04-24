package controller

import (
	"net/http"

	"github.com/Rizal-Nurochman/matchnbuild/modules/designer/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/designer/service"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/utils"
	"github.com/gin-gonic/gin"
)

type (
	designerProfileHandler struct {
		designerService service.DesignerService
	}

	DesignerProfileHandler interface {
		GetAll(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		GetMyProfile(ctx *gin.Context)
		Update(ctx *gin.Context)
	}
)

func NewDesignerProfileHandler(designerService service.DesignerService) DesignerProfileHandler  {
	return &designerProfileHandler{designerService: designerService}
}

func (h *designerProfileHandler) GetAll(ctx *gin.Context)  {
	designers, err := h.designerService.GetAll(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DESIGNER_PROFILE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_DESIGNER_PROFILE, designers)
	ctx.JSON(http.StatusOK, res)
}

func (h *designerProfileHandler) GetByID(ctx *gin.Context)  {
	id := ctx.Param("id")
	
	designer, err := h.designerService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DESIGNER_PROFILE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res:=utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_DESIGNER_PROFILE, designer)
	ctx.JSON(http.StatusOK, res)
}

func (h *designerProfileHandler) GetMyProfile(ctx *gin.Context)  {
	
}

func (h *designerProfileHandler) Update(ctx *gin.Context)  {
	
}