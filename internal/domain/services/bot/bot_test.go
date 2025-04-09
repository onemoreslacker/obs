package bot_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/commands"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/services/bot"
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
			want:   bot.UnknownCommand,
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

			cfg, _ := config.Load("config/config.yaml")

			b, _ := bot.New(scrapperClient, api, cfg)

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
