package service

import (
	"context"
	"fmt"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	chatDto "github.com/Rizal-Nurochman/matchnbuild/modules/chat/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/chat/repository"
	prRepo "github.com/Rizal-Nurochman/matchnbuild/modules/project_request/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatService interface {
	GetConversations(ctx context.Context, userID string) ([]chatDto.ConversationResponse, error)
	GetMessages(ctx context.Context, userID string, conversationID string, beforeID string, limit int) (chatDto.GetMessagesResponse, error)
	SendMessage(ctx context.Context, userID string, conversationID string, req chatDto.SendMessageRequest) (chatDto.MessageResponse, error)
}

type chatService struct {
	messageRepo                repository.MessageRepository
	conversationParticipantRepo prRepo.ConversationParticipantRepository
	conversationRepo           prRepo.ConversationRepository
	db                         *gorm.DB
}

func NewChatService(
	messageRepo repository.MessageRepository,
	conversationParticipantRepo prRepo.ConversationParticipantRepository,
	conversationRepo prRepo.ConversationRepository,
	db *gorm.DB,
) ChatService {
	return &chatService{
		messageRepo:                messageRepo,
		conversationParticipantRepo: conversationParticipantRepo,
		conversationRepo:           conversationRepo,
		db:                         db,
	}
}

func (s *chatService) GetConversations(ctx context.Context, userID string) ([]chatDto.ConversationResponse, error) {
	participants, err := s.conversationParticipantRepo.GetByUserID(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}

	var responses []chatDto.ConversationResponse
	for _, p := range participants {
		conv, err := s.conversationRepo.GetByID(ctx, s.db, p.ConversationID.String())
		if err != nil {
			continue
		}

		resp := chatDto.ConversationResponse{
			ID:               conv.ID.String(),
			ProjectRequestID: conv.ProjectRequestID.String(),
			CreatedAt:        conv.CreatedAt,
		}
		if conv.OrderID != nil {
			orderID := conv.OrderID.String()
			resp.OrderID = &orderID
		}

		responses = append(responses, resp)
	}

	if responses == nil {
		responses = []chatDto.ConversationResponse{}
	}

	return responses, nil
}

func (s *chatService) GetMessages(ctx context.Context, userID string, conversationID string, beforeID string, limit int) (chatDto.GetMessagesResponse, error) {
	isParticipant, err := s.conversationParticipantRepo.IsParticipant(ctx, s.db, conversationID, userID)
	if err != nil {
		return chatDto.GetMessagesResponse{}, err
	}
	if !isParticipant {
		return chatDto.GetMessagesResponse{}, chatDto.ErrNotConversationParticipant
	}

	if limit <= 0 || limit > 50 {
		limit = 30
	}

	messages, err := s.messageRepo.GetByConversationID(ctx, s.db, conversationID, beforeID, limit+1)
	if err != nil {
		return chatDto.GetMessagesResponse{}, err
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	responses := make([]chatDto.MessageResponse, len(messages))
	for i, msg := range messages {
		responses[i] = toMessageResponse(msg)
	}

	return chatDto.GetMessagesResponse{
		Messages: responses,
		HasMore:  hasMore,
	}, nil
}

func (s *chatService) SendMessage(ctx context.Context, userID string, conversationID string, req chatDto.SendMessageRequest) (chatDto.MessageResponse, error) {
	isParticipant, err := s.conversationParticipantRepo.IsParticipant(ctx, s.db, conversationID, userID)
	if err != nil {
		return chatDto.MessageResponse{}, err
	}
	if !isParticipant {
		return chatDto.MessageResponse{}, chatDto.ErrNotConversationParticipant
	}

	senderUUID, err := uuid.Parse(userID)
	if err != nil {
		return chatDto.MessageResponse{}, fmt.Errorf("invalid user id: %w", err)
	}

	convUUID, err := uuid.Parse(conversationID)
	if err != nil {
		return chatDto.MessageResponse{}, fmt.Errorf("invalid conversation id: %w", err)
	}

	message := entities.Message{
		ID:              uuid.New(),
		ConversationID:  convUUID,
		SenderID:        senderUUID,
		ClientMessageID: &req.ClientMessageID,
		MessageText:     req.MessageText,
		AttachmentURL:   req.AttachmentURL,
		MessageType:     req.MessageType,
	}

	savedMessage, err := s.messageRepo.Create(ctx, s.db, message)
	if err != nil {
		return chatDto.MessageResponse{}, err
	}

	savedMessage, err = s.messageRepo.GetByID(ctx, s.db, savedMessage.ID.String())
	if err != nil {
		return chatDto.MessageResponse{}, err
	}

	return toMessageResponse(savedMessage), nil
}

func toMessageResponse(msg entities.Message) chatDto.MessageResponse {
	resp := chatDto.MessageResponse{
		ID:             msg.ID.String(),
		ConversationID: msg.ConversationID.String(),
		SenderID:       msg.SenderID.String(),
		SenderName:     msg.Sender.Name,
		MessageText:    msg.MessageText,
		AttachmentURL:  msg.AttachmentURL,
		MessageType:    msg.MessageType,
		CreatedAt:      msg.CreatedAt,
	}
	if msg.ClientMessageID != nil {
		resp.ClientMessageID = msg.ClientMessageID
	}
	return resp
}
