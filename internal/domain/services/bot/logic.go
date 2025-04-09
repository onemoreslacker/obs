package bot

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/commands"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) isRegistered(chatID int64) error {
	resp, err := b.scrapperClient.GetTgChatId(context.Background(), chatID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return ErrUserNotRegistered
	}

	var chatIDs []int64

	if err := json.NewDecoder(resp.Body).Decode(&chatIDs); err != nil {
		return ErrUserNotRegistered
	}

	slog.Info(
		"Bot: registered chats",
		slog.Any("chatIDs", chatIDs),
	)

	for _, id := range chatIDs {
		if chatID == id {
			return nil
		}
	}

	return ErrUserNotRegistered
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
