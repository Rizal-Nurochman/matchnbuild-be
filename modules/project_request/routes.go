package project_request

import (
	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	"github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	"github.com/Rizal-Nurochman/matchnbuild/modules/project_request/controller"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	prController := do.MustInvoke[controller.ProjectRequestController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	prRoutes := server.Group("/project-request")
	{
		prRoutes.POST("", middlewares.Authenticate(jwtService), prController.Create)
		prRoutes.GET("/:id", middlewares.Authenticate(jwtService), prController.GetByID)
		prRoutes.GET("/my-requests", middlewares.Authenticate(jwtService), prController.GetMyRequests)
		prRoutes.GET("/incoming", middlewares.Authenticate(jwtService), prController.GetMyIncomingRequests)
	}
}
