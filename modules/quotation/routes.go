package quotation

import (
	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	authService "github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	chatRepo "github.com/Rizal-Nurochman/matchnbuild/modules/chat/repository"
	chatService "github.com/Rizal-Nurochman/matchnbuild/modules/chat/service"
	chatWs "github.com/Rizal-Nurochman/matchnbuild/modules/chat/websocket"
	prRepo "github.com/Rizal-Nurochman/matchnbuild/modules/project_request/repository"
	"github.com/Rizal-Nurochman/matchnbuild/modules/quotation/controller"
	qRepo "github.com/Rizal-Nurochman/matchnbuild/modules/quotation/repository"
	qService "github.com/Rizal-Nurochman/matchnbuild/modules/quotation/service"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)
	hub := do.MustInvokeNamed[*chatWs.Hub](injector, constants.ChatHub)

	// Repositories
	quotationRepository := qRepo.NewQuotationRepository(db)
	orderRepository := qRepo.NewOrderRepository(db)
	projectRequestRepository := prRepo.NewProjectRequestRepository(db)
	designerProfileRepository := prRepo.NewDesignerProfileRepository(db)
	conversationRepository := prRepo.NewConversationRepository(db)
	conversationParticipantRepository := prRepo.NewConversationParticipantRepository(db)
	messageRepository := chatRepo.NewMessageRepository(db)

	// Chat service lokal untuk broadcast notif quotation
	chatSvc := chatService.NewChatService(messageRepository, conversationParticipantRepository, conversationRepository, db)
	chatSvc.SetBroadcaster(hub)

	// Quotation service + controller
	qSvc := qService.NewQuotationService(quotationRepository, orderRepository, projectRequestRepository, designerProfileRepository, conversationRepository, chatSvc, db)
	qCtrl := controller.NewQuotationController(qSvc)

	qRoutes := server.Group("/quotation")
	{
		qRoutes.POST("", middlewares.Authenticate(jwtService), qCtrl.Create)
		qRoutes.GET("/:id", middlewares.Authenticate(jwtService), qCtrl.GetByID)
		qRoutes.GET("/project-request/:project_request_id", middlewares.Authenticate(jwtService), qCtrl.GetByProjectRequestID)
		qRoutes.PUT("/:id/accept", middlewares.Authenticate(jwtService), qCtrl.Accept)
		qRoutes.PUT("/:id/reject", middlewares.Authenticate(jwtService), qCtrl.Reject)
	}
}
