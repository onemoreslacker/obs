package telebot_test

import (
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/commands"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/telebot"

	"github.com/stretchr/testify/require"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestUnknownCommand(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		chatID int64
		want   string
	}{
		{
			name:   "unknown command",
			msg:    "/kill",
			chatID: 1,
			want:   telebot.UnknownCommand,
		},
		{
			name:   "track command",
			msg:    "/track",
			chatID: 1,
			want:   commands.TrackRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := NewMockTgAPI(t)
			scrapperClient := NewMockScrapperClient(t)

			b, _ := telebot.New(scrapperClient, api)

			reply := b.InitializeCommand(newTestCommand(tt.msg, tt.chatID))

			require.Equal(t, tt.want, reply.Text)
		})
	}
}

func newTestCommand(text string, chatID int64) *tgbotapi.Message {
	return &tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: chatID},
		Text: text,
		Entities: []tgbotapi.MessageEntity{
			{
				Type:   "bot_command",
				Offset: 0,
				Length: len(text),
			},
		},
	}
}
