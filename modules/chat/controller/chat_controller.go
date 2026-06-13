package controller

import (
	"net/http"
	"strconv"

	"github.com/Rizal-Nurochman/matchnbuild/modules/chat/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/chat/service"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/utils"
	"github.com/gin-gonic/gin"
)

type (
	ChatController interface {
		GetConversations(ctx *gin.Context)
		GetMessages(ctx *gin.Context)
		SendMessage(ctx *gin.Context)
		GetUnreadCount(ctx *gin.Context)
		GetTotalUnreadCount(ctx *gin.Context)
	}

	chatController struct {
		chatService service.ChatService
	}
)

func NewChatController(cs service.ChatService) ChatController {
	return &chatController{
		chatService: cs,
	}
}

func (c *chatController) GetConversations(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)

	result, err := c.chatService.GetConversations(ctx.Request.Context(), userID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_CONVERSATIONS, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_CONVERSATIONS, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *chatController) GetMessages(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	conversationID := ctx.Param("id")

	beforeID := ctx.Query("before")
	afterID := ctx.Query("after")
	limitStr := ctx.DefaultQuery("limit", "30")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 30
	}

	result, err := c.chatService.GetMessages(ctx.Request.Context(), userID, conversationID, beforeID, afterID, limit)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == dto.ErrNotConversationParticipant {
			statusCode = http.StatusForbidden
		}
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_MESSAGES, err.Error(), nil)
		ctx.JSON(statusCode, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_MESSAGES, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *chatController) SendMessage(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	conversationID := ctx.Param("id")

	var req dto.SendMessageRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed("failed to get data from body", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.chatService.SendMessage(ctx.Request.Context(), userID, conversationID, req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == dto.ErrNotConversationParticipant {
			statusCode = http.StatusForbidden
		}
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_SEND_MESSAGE, err.Error(), nil)
		ctx.JSON(statusCode, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_SEND_MESSAGE, result)
	ctx.JSON(http.StatusCreated, res)
}

func (c *chatController) GetUnreadCount(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	conversationID := ctx.Param("id")

	count, err := c.chatService.GetUnreadCount(userID, conversationID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_UNREAD_COUNT, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_UNREAD_COUNT, dto.UnreadCountResponse{
		ConversationID: conversationID,
		UnreadCount:    int(count),
	})
	ctx.JSON(http.StatusOK, res)
}

func (c *chatController) GetTotalUnreadCount(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)

	count, err := c.chatService.GetTotalUnreadCount(userID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_UNREAD_COUNT, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_UNREAD_COUNT, dto.TotalUnreadCountResponse{
		TotalUnreadCount: int(count),
	})
	ctx.JSON(http.StatusOK, res)
}
