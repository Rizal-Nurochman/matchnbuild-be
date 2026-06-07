package payment

import (
	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	authService "github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	"github.com/Rizal-Nurochman/matchnbuild/modules/payment/controller"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	paymentController := do.MustInvoke[controller.PaymentController](injector)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	routes := server.Group("/payment")
	{
		routes.POST("/notification", paymentController.Notification)
		routes.POST("/:order_id/snap-token", middlewares.Authenticate(jwtService), paymentController.CreateSnapToken)
		routes.GET("/:order_id/status", middlewares.Authenticate(jwtService), paymentController.GetStatus)
	}
}
