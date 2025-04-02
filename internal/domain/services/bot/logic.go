package bot

import (
	"context"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/commands"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) isRegistered(id int64) error {
	resp, err := b.scrapperClient.GetTgChatId(context.Background(), id)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return ErrUserNotRegistered
	}

	return nil
}

func (b *Bot) InitializeCommand(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var reply tgbotapi.MessageConfig

	switch msg.Command() {
	case track:
		b.currentCommand = commands.NewCommandTrack(
			msg.Chat.ID,
			b.scrapperClient,
		)
	case untrack:
		b.currentCommand = commands.NewCommandUntrack(
			msg.Chat.ID,
			b.scrapperClient,
		)
	case list:
		b.currentCommand = commands.NewCommandList(
			msg.Chat.ID,
			b.scrapperClient,
		)
	default:
		return tgbotapi.NewMessage(
			msg.Chat.ID,
			UnknownCommand,
		)
	}

	reply = b.configureReply(msg)

	return reply
}

const (
	UnknownCommand = "⚡️ Unknown command! Use /help to get a list of available commands."
)
