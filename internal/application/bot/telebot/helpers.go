package telebot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	commands2 "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/commands"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) configureUpdates() tgbotapi.UpdatesChannel {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := b.tgb.GetUpdatesChan(updateConfig)

	time.Sleep(250 * time.Millisecond)
	updates.Clear()

	return updates
}

// helpCommand constructs available commands message.
func constructHelpMessage() string {
	var msg string
	for _, cmd := range config.Descriptions {
		msg += fmt.Sprintf("%s - %s\n", cmd.Name, cmd.Description)
	}

	return msg
}

func (b *Bot) configureReply(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	content, keyboard := b.currentCommand.Stage()

	reply := tgbotapi.NewMessage(msg.Chat.ID, content)
	if keyboard {
		reply.ReplyMarkup = inlineKeyboard
	}

	return reply
}

func (b *Bot) withAuthorization(id int64, next func() string) string {
	if err := b.isRegistered(id); err != nil {
		return Unregistered
	}

	return next()
}

var inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Yes", "Yes"),
		tgbotapi.NewInlineKeyboardButtonData("No", "No"),
	),
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
		b.currentCommand = commands2.NewCommandTrack(
			msg.Chat.ID,
			b.scrapperClient,
		)
	case untrack:
		b.currentCommand = commands2.NewCommandUntrack(
			msg.Chat.ID,
			b.scrapperClient,
		)
	case list:
		b.currentCommand = commands2.NewCommandList(
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

const (
	Unregistered = "⚡ You are not registered! Press /start to register!"
)
