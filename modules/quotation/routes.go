package quotation

import (
	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	"github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	"github.com/Rizal-Nurochman/matchnbuild/modules/quotation/controller"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	qController := do.MustInvoke[controller.QuotationController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	qRoutes := server.Group("/quotation")
	{
		qRoutes.POST("", middlewares.Authenticate(jwtService), qController.Create)
		qRoutes.GET("/:id", middlewares.Authenticate(jwtService), qController.GetByID)
		qRoutes.PUT("/:id/accept", middlewares.Authenticate(jwtService), qController.Accept)
		qRoutes.PUT("/:id/reject", middlewares.Authenticate(jwtService), qController.Reject)
	}
}
