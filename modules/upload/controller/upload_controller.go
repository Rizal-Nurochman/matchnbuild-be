package controller

import (
	"net/http"

	"github.com/Rizal-Nurochman/matchnbuild/pkg/utils"
	"github.com/gin-gonic/gin"
)

type (
	UploadController interface {
		UploadFile(ctx *gin.Context)
	}

	uploadController struct{}
)

func NewUploadController() UploadController {
	return &uploadController{}
}

func (c *uploadController) UploadFile(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		res := utils.BuildResponseFailed("failed to get file from request", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	defer file.Close()

	folder := ctx.DefaultPostForm("folder", "general")

	url, err := utils.UploadToImageKit(file, header, folder)
	if err != nil {
		res := utils.BuildResponseFailed("failed to upload file", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("success upload file", gin.H{
		"url": url,
	})
	ctx.JSON(http.StatusOK, res)
}
