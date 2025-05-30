package telebot_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/commands"
	mocks "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/telebot"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestInitializeCommand(t *testing.T) {
	var chatID int64 = 1

	tests := map[string]struct {
		msg           *tgbotapi.Message
		chatID        int64
		expectedReply string
	}{
		"unknown command": {
			msg: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				Text: "/kill",
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: len("/kill"),
					},
				},
			},
			expectedReply: telebot.UnknownCommand,
		},
		"track command": {
			msg: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				Text: "/track",
				Entities: []tgbotapi.MessageEntity{
					{
						Type:   "bot_command",
						Offset: 0,
						Length: len("/track"),
					},
				},
			},
			expectedReply: commands.TrackRequest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			api := mocks.NewMockTgAPI(t)
			client := mocks.NewMockScrapperClient(t)

			b := &telebot.Bot{
				Client: client,
				Tgb:    api,
			}

			actualReply := b.InitializeCommand(test.msg)
			require.Equal(t, test.expectedReply, actualReply.Text)
		})
	}
}

func TestIsRegistered(t *testing.T) {
	tests := map[string]struct {
		chatID     int64
		statusCode int
		wantErr    bool
	}{
		"user is registered": {
			chatID:     1,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		"user is not registered": {
			chatID:     1,
			statusCode: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := mocks.NewMockScrapperClient(t)
			defer client.AssertExpectations(t)

			client.On("GetTgChatId", mock.Anything, test.chatID).
				Once().Return(&http.Response{
				StatusCode: test.statusCode,
				Body:       io.NopCloser(bytes.NewReader(nil)),
			}, nil)

			b := &telebot.Bot{
				Client: client,
			}
			err := b.IsRegistered(test.chatID)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
