package upload

import (
	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	"github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	"github.com/Rizal-Nurochman/matchnbuild/modules/upload/controller"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	uploadController := controller.NewUploadController()
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	uploadRoutes := server.Group("/upload")
	{
		uploadRoutes.POST("", middlewares.Authenticate(jwtService), uploadController.UploadFile)
	}
}
