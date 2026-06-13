package chat

import (
	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	chatController "github.com/Rizal-Nurochman/matchnbuild/modules/chat/controller"
	chatRepo "github.com/Rizal-Nurochman/matchnbuild/modules/chat/repository"
	chatService "github.com/Rizal-Nurochman/matchnbuild/modules/chat/service"
	authService "github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	prRepo "github.com/Rizal-Nurochman/matchnbuild/modules/project_request/repository"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	messageRepo := chatRepo.NewMessageRepository(db)
	conversationRepo := prRepo.NewConversationRepository(db)
	conversationParticipantRepo := prRepo.NewConversationParticipantRepository(db)
	chatSvc := chatService.NewChatService(messageRepo, conversationParticipantRepo, conversationRepo, db)
	chatCtrl := chatController.NewChatController(chatSvc)

	chatRoutes := server.Group("/conversations")
	{
		chatRoutes.GET("", middlewares.Authenticate(jwtService), chatCtrl.GetConversations)
		chatRoutes.GET("/:id/messages", middlewares.Authenticate(jwtService), chatCtrl.GetMessages)
		chatRoutes.POST("/:id/messages", middlewares.Authenticate(jwtService), chatCtrl.SendMessage)
	}
}
