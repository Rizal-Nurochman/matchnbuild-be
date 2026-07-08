package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	chatDto "github.com/Rizal-Nurochman/matchnbuild/modules/chat/dto"
	"github.com/Rizal-Nurochman/matchnbuild/modules/chat/repository"
	prRepo "github.com/Rizal-Nurochman/matchnbuild/modules/project_request/repository"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const wsOperationTimeout = 5 * time.Second

type Broadcaster interface {
	BroadcastToRoomExcept(roomID string, eventType string, data interface{}, excludeUserID string)
}

type ChatService interface {
	GetConversations(ctx context.Context, userID string) ([]chatDto.ConversationResponse, error)
	GetMessages(ctx context.Context, userID string, conversationID string, beforeID string, afterID string, limit int) (chatDto.GetMessagesResponse, error)
	SendMessage(ctx context.Context, userID string, conversationID string, req chatDto.SendMessageRequest) (chatDto.MessageResponse, error)
	IsParticipant(userID string, conversationID string) (bool, error)
	MarkAsRead(userID string, conversationID string, messageID string) error
	SendMessageWS(userID string, conversationID string, req chatDto.SendMessageRequest) (chatDto.MessageResponse, error)
	GetUnreadCount(userID string, conversationID string) (int64, error)
	GetTotalUnreadCount(userID string) (int64, error)
	SetBroadcaster(b Broadcaster)
}

type chatService struct {
	messageRepo                repository.MessageRepository
	conversationParticipantRepo prRepo.ConversationParticipantRepository
	conversationRepo           prRepo.ConversationRepository
	db                         *gorm.DB
	broadcaster                Broadcaster
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

func (s *chatService) SetBroadcaster(b Broadcaster) {
	s.broadcaster = b
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

		pr := conv.ProjectRequest
		resp := chatDto.ConversationResponse{
			ID:               conv.ID.String(),
			ProjectRequestID: conv.ProjectRequestID.String(),
			CreatedAt:        conv.CreatedAt,
			ProjectRequest: chatDto.ConversationProjectRequestInfo{
				ID:            pr.ID.String(),
				Description:   pr.Description,
				Status:        pr.Status,
				InitialBudget: pr.InitialBudget.InexactFloat64(),
			},
			Client: chatDto.ConversationUserInfo{
				ID:             pr.Client.ID.String(),
				Name:           pr.Client.Name,
				ProfilePicture: pr.Client.ProfilePicture,
			},
			Designer: chatDto.ConversationUserInfo{
				ID:             pr.Designer.UserID.String(),
				Name:           pr.Designer.User.Name,
				ProfilePicture: pr.Designer.User.ProfilePicture,
			},
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

func (s *chatService) GetMessages(ctx context.Context, userID string, conversationID string, beforeID string, afterID string, limit int) (chatDto.GetMessagesResponse, error) {
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

	messages, err := s.messageRepo.GetByConversationID(ctx, s.db, conversationID, beforeID, afterID, limit+1)
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
	normalized, err := validateSendMessage(req)
	if err != nil {
		return chatDto.MessageResponse{}, err
	}
	req = normalized

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
		if isDuplicateKeyErr(err) {
			existing, getErr := s.messageRepo.GetBySenderAndClientMessageID(ctx, s.db, userID, req.ClientMessageID)
			if getErr == nil {
				return toMessageResponse(existing), nil
			}
		}
		return chatDto.MessageResponse{}, err
	}

	savedMessage, err = s.messageRepo.GetByID(ctx, s.db, savedMessage.ID.String())
	if err != nil {
		return chatDto.MessageResponse{}, err
	}

	resp := toMessageResponse(savedMessage)
	s.broadcastMessageCreated(resp)
	return resp, nil
}

func (s *chatService) broadcastMessageCreated(resp chatDto.MessageResponse) {
	if s.broadcaster == nil {
		return
	}

	created := chatDto.MessageCreatedData{
		ID:             resp.ID,
		ConversationID: resp.ConversationID,
		SenderID:       resp.SenderID,
		SenderName:     resp.SenderName,
		ClientMessageID: resp.ClientMessageID,
		MessageType:    resp.MessageType,
		Text:           resp.MessageText,
		AttachmentURL:  resp.AttachmentURL,
		CreatedAt:      resp.CreatedAt,
	}
	s.broadcaster.BroadcastToRoomExcept(resp.ConversationID, chatDto.EVENT_MESSAGE_CREATED, created, resp.SenderID)
}

func (s *chatService) IsParticipant(userID string, conversationID string) (bool, error) {
	ctx, cancel := wsContext()
	defer cancel()
	return s.conversationParticipantRepo.IsParticipant(ctx, s.db, conversationID, userID)
}

func (s *chatService) MarkAsRead(userID string, conversationID string, messageID string) error {
	ctx, cancel := wsContext()
	defer cancel()

	msg, err := s.messageRepo.GetByID(ctx, s.db, messageID)
	if err != nil {
		return chatDto.ErrMessageNotFound
	}
	if msg.ConversationID.String() != conversationID {
		return chatDto.ErrMessageNotFound
	}

	return s.conversationParticipantRepo.UpdateLastReadMessage(ctx, s.db, conversationID, userID, messageID)
}

func (s *chatService) SendMessageWS(userID string, conversationID string, req chatDto.SendMessageRequest) (chatDto.MessageResponse, error) {
	ctx, cancel := wsContext()
	defer cancel()
	return s.SendMessage(ctx, userID, conversationID, req)
}

func (s *chatService) GetUnreadCount(userID string, conversationID string) (int64, error) {
	ctx, cancel := wsContext()
	defer cancel()
	return s.messageRepo.GetUnreadCount(ctx, s.db, conversationID, userID)
}

func (s *chatService) GetTotalUnreadCount(userID string) (int64, error) {
	ctx, cancel := wsContext()
	defer cancel()
	return s.messageRepo.GetTotalUnreadCount(ctx, s.db, userID)
}

func wsContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), wsOperationTimeout)
}

func validateSendMessage(req chatDto.SendMessageRequest) (chatDto.SendMessageRequest, error) {
	req.MessageText = strings.TrimSpace(req.MessageText)
	req.AttachmentURL = strings.TrimSpace(req.AttachmentURL)

	if req.MessageType == "" {
		req.MessageType = constants.MESSAGE_TYPE_TEXT
	}
	switch req.MessageType {
	case constants.MESSAGE_TYPE_TEXT, constants.MESSAGE_TYPE_IMAGE, constants.MESSAGE_TYPE_FILE:
	default:
		return req, chatDto.ErrInvalidMessageType
	}

	if req.MessageText == "" && req.AttachmentURL == "" {
		return req, chatDto.ErrEmptyMessage
	}

	if len(req.MessageText) > chatDto.MaxMessageTextLength {
		return req, chatDto.ErrMessageTextTooLong
	}
	if len(req.AttachmentURL) > chatDto.MaxAttachmentURLLength {
		return req, chatDto.ErrAttachmentURLTooLong
	}
	if len(req.ClientMessageID) > chatDto.MaxClientMessageIDLen {
		return req, chatDto.ErrClientMessageIDTooLong
	}

	return req, nil
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

func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return strings.Contains(err.Error(), "23505") ||
		strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
