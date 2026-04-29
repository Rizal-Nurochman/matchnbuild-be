package designer

import (
	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	"github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	"github.com/Rizal-Nurochman/matchnbuild/modules/designer/controller"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	designerHandler := do.MustInvoke[controller.DesignerProfileHandler](injector)
	JwtSvc := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)
	DesignerRouter := server.Group("/designers")
	{
		DesignerRouter.GET("", designerHandler.GetAll)
		DesignerRouter.GET("/:id", designerHandler.GetByID)
		DesignerRouter.GET("/me", middlewares.Authenticate(JwtSvc), designerHandler.GetMyProfile)
		DesignerRouter.PATCH("/:id", middlewares.Authenticate(JwtSvc), designerHandler.Update)
	}
}