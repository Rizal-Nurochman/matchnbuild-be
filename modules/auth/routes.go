package auth

import (
	"github.com/Rizal-Nurochman/matchnbuild/modules/auth/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	authController := do.MustInvoke[controller.AuthController](injector)

	authRoutes := server.Group("/auth")
	{
		authRoutes.POST("/register", authController.Register)
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/refresh", authController.RefreshToken)
		authRoutes.POST("/logout", authController.Logout)
		authRoutes.POST("/send-verification-email", authController.SendVerificationEmail)
		authRoutes.POST("/verify-email", authController.VerifyEmail)
		authRoutes.POST("/send-password-reset", authController.SendPasswordReset)
		authRoutes.POST("/reset-password", authController.ResetPassword)
	}
}
