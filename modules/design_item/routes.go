package design_item

import (
	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	"github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	"github.com/Rizal-Nurochman/matchnbuild/modules/design_item/controller"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	designItemHandler := do.MustInvoke[controller.DesignItemHandler](injector)
	JwtSvc := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	designItemRouter := server.Group("/design-items")
	{
		designItemRouter.GET("", designItemHandler.GetAll)
		designItemRouter.GET("/my-items", middlewares.Authenticate(JwtSvc), designItemHandler.GetMyItems)
		designItemRouter.GET("/:id", designItemHandler.GetByID)
		designItemRouter.POST("", middlewares.Authenticate(JwtSvc), designItemHandler.Create)
		designItemRouter.PATCH("/:id", middlewares.Authenticate(JwtSvc), designItemHandler.Update)
		designItemRouter.DELETE("/:id", middlewares.Authenticate(JwtSvc), designItemHandler.Delete)
	}

	featuresRouter := server.Group("/features")
	{
		featuresRouter.GET("", designItemHandler.GetAllFeatures)
	}
}