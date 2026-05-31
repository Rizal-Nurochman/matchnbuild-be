package recommendation

import (
	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	"github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	"github.com/Rizal-Nurochman/matchnbuild/modules/recommendation/controller"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	handler := do.MustInvoke[controller.RecommendationHandler](injector)
	jwtSvc := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	recRouter := server.Group("/recommendations")
	recRouter.Use(middlewares.Authenticate(jwtSvc))
	{
		recRouter.POST("", handler.GetRecommendations)
	}
}