package telebot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/commands"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) ConfigureUpdates() tgbotapi.UpdatesChannel {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := b.Tgb.GetUpdatesChan(updateConfig)

	updates.Clear()

	return updates
}

func ConstructHelpMessage() string {
	var msg string
	for _, cmd := range config.Descriptions {
		msg += fmt.Sprintf("%s - %s\n", cmd.Name, cmd.Description)
	}

	return msg
}

func (b *Bot) ConfigureReply(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	content, keyboard := b.CurrentCommand.Stage()

	reply := tgbotapi.NewMessage(msg.Chat.ID, content)
	if keyboard {
		reply.ReplyMarkup = inlineKeyboard
	}

	return reply
}

var inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Yes", "Yes"),
		tgbotapi.NewInlineKeyboardButtonData("No", "No"),
	),
)

func (b *Bot) IsRegistered(chatID int64) error {
	resp, err := b.Client.GetTgChatId(context.Background(), chatID)
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
	case Track:
		b.CurrentCommand = commands.NewTrack(msg.Chat.ID, b.Client)
	case Untrack:
		b.CurrentCommand = commands.NewUntrack(msg.Chat.ID, b.Client)
	case List:
		b.CurrentCommand = commands.NewList(msg.Chat.ID, b.Client)
	default:
		return tgbotapi.NewMessage(msg.Chat.ID, UnknownCommand)
	}

	reply = b.ConfigureReply(msg)

	return reply
}

const (
	UnknownCommand = "⚡️ Unknown command! Use /help to get a list of available commands."
)
