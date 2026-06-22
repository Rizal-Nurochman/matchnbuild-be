package service

import (
	"testing"

	chatDto "github.com/Rizal-Nurochman/matchnbuild/modules/chat/dto"
)

func TestValidateSendMessage(t *testing.T) {
	long := func(n int) string {
		b := make([]byte, n)
		for i := range b {
			b[i] = 'a'
		}
		return string(b)
	}

	tests := []struct {
		name    string
		in      chatDto.SendMessageRequest
		wantErr error
		wantType string
		wantText string
	}{
		{
			name:     "defaults to Text type and trims",
			in:       chatDto.SendMessageRequest{ClientMessageID: "c1", MessageText: "  hi  "},
			wantErr:  nil,
			wantType: "Text",
			wantText: "hi",
		},
		{
			name:    "rejects unknown message type",
			in:      chatDto.SendMessageRequest{ClientMessageID: "c1", MessageText: "hi", MessageType: "Sticker"},
			wantErr: chatDto.ErrInvalidMessageType,
		},
		{
			name:    "rejects empty text and attachment",
			in:      chatDto.SendMessageRequest{ClientMessageID: "c1", MessageText: "   ", MessageType: "Text"},
			wantErr: chatDto.ErrEmptyMessage,
		},
		{
			name:     "allows attachment-only message",
			in:       chatDto.SendMessageRequest{ClientMessageID: "c1", AttachmentURL: "https://x/y.png", MessageType: "Image"},
			wantErr:  nil,
			wantType: "Image",
		},
		{
			name:    "rejects oversized text",
			in:      chatDto.SendMessageRequest{ClientMessageID: "c1", MessageText: long(chatDto.MaxMessageTextLength + 1), MessageType: "Text"},
			wantErr: chatDto.ErrMessageTextTooLong,
		},
		{
			name:    "rejects oversized attachment url",
			in:      chatDto.SendMessageRequest{ClientMessageID: "c1", MessageText: "hi", AttachmentURL: long(chatDto.MaxAttachmentURLLength + 1), MessageType: "Text"},
			wantErr: chatDto.ErrAttachmentURLTooLong,
		},
		{
			name:    "rejects oversized client_message_id",
			in:      chatDto.SendMessageRequest{ClientMessageID: long(chatDto.MaxClientMessageIDLen + 1), MessageText: "hi", MessageType: "Text"},
			wantErr: chatDto.ErrClientMessageIDTooLong,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateSendMessage(tc.in)
			if err != tc.wantErr {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			if tc.wantType != "" && got.MessageType != tc.wantType {
				t.Errorf("MessageType = %q, want %q", got.MessageType, tc.wantType)
			}
			if tc.wantText != "" && got.MessageText != tc.wantText {
				t.Errorf("MessageText = %q, want %q", got.MessageText, tc.wantText)
			}
		})
	}
}
