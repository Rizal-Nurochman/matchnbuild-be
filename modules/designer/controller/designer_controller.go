package controller

import (
	"net/http"

	"github.com/Caknoooo/go-pagination"
	"github.com/Rizal-Nurochman/matchnbuild/modules/designer/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/designer/query"
	"github.com/Rizal-Nurochman/matchnbuild/modules/designer/service"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	designerProfileHandler struct {
		designerService service.DesignerService
		db *gorm.DB
	}

	DesignerProfileHandler interface {
		GetAll(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		GetMyProfile(ctx *gin.Context)
		Update(ctx *gin.Context)
	}
)

func NewDesignerProfileHandler(injector *do.Injector, designerService service.DesignerService) DesignerProfileHandler  {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &designerProfileHandler{designerService: designerService, db: db}
}

func (h *designerProfileHandler) GetAll(ctx *gin.Context)  {
	var filter = &query.DesignerFilter{}
	filter.BindPagination(ctx)

	ctx.ShouldBindQuery(filter)

	designers, total, err := pagination.PaginatedQueryWithIncludable[query.Designer](h.db, filter)
	if err != nil {
		res:=utils.BuildResponseFailed(dto.ErrDesignerProfileNotFound.Error(), err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	paginationRes := pagination.CalculatePagination(filter.Pagination, total)
	res := pagination.NewPaginatedResponse(http.StatusOK, dto.MESSAGE_SUCCESS_GET_DESIGNER_PROFILE, designers, paginationRes)
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
	userID := ctx.MustGet("user_id").(string)

	designer, err := h.designerService.GetMyProfile(ctx.Request.Context(), userID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DESIGNER_PROFILE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_DESIGNER_PROFILE, designer)
	ctx.JSON(http.StatusOK, res)
}

func (h *designerProfileHandler) Update(ctx *gin.Context)  {
	var req dto.DesignerProfileUpdateRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_DESIGNER_PROFILE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}


}