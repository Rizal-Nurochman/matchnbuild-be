package user_preferences

import (
	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	"github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	"github.com/Rizal-Nurochman/matchnbuild/modules/user_preferences/controller"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	userPreferenceHandler := do.MustInvoke[controller.UserPreferenceHandler](injector)
	JwtSvc := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	userPreferenceRouter := server.Group("/user-preferences")
	{
		userPreferenceRouter.GET("/:id", middlewares.Authenticate(JwtSvc), userPreferenceHandler.GetByID)
		userPreferenceRouter.POST("", middlewares.Authenticate(JwtSvc), userPreferenceHandler.Create)
		userPreferenceRouter.PATCH("/:id", middlewares.Authenticate(JwtSvc), userPreferenceHandler.Update)
	}
}