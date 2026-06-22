package chat

import (
	"log"
	"os"

	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	chatController "github.com/Rizal-Nurochman/matchnbuild/modules/chat/controller"
	chatRepo "github.com/Rizal-Nurochman/matchnbuild/modules/chat/repository"
	chatService "github.com/Rizal-Nurochman/matchnbuild/modules/chat/service"
	chatWs "github.com/Rizal-Nurochman/matchnbuild/modules/chat/websocket"
	authService "github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	userRepo "github.com/Rizal-Nurochman/matchnbuild/modules/user/repository"
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

	userRepository := userRepo.NewUserRepository(db)

	allowedOrigin := os.Getenv("WS_ALLOWED_ORIGIN")

	hub := chatWs.NewHub(nil)
	go hub.Run()
	log.Println("[WS] Hub started")

	// Register the hub in the DI container so its lifecycle is managed centrally
	// (graceful shutdown on server stop) and other modules can reach it.
	do.ProvideNamedValue(injector, constants.ChatHub, hub)

	// Let the service broadcast created messages so REST and WS share one path.
	chatSvc.SetBroadcaster(hub)

	wsHandler := chatWs.NewHandler(hub, chatSvc, userRepository, jwtService, db, allowedOrigin)

	chatRoutes := server.Group("/conversations")
	{
		chatRoutes.GET("", middlewares.Authenticate(jwtService), chatCtrl.GetConversations)
		chatRoutes.GET("/:id/messages", middlewares.Authenticate(jwtService), chatCtrl.GetMessages)
		chatRoutes.POST("/:id/messages", middlewares.Authenticate(jwtService), chatCtrl.SendMessage)
		chatRoutes.GET("/:id/unread-count", middlewares.Authenticate(jwtService), chatCtrl.GetUnreadCount)
		chatRoutes.GET("/unread-count", middlewares.Authenticate(jwtService), chatCtrl.GetTotalUnreadCount)
	}

	server.GET("/ws", wsHandler.HandleWebSocket)
}
