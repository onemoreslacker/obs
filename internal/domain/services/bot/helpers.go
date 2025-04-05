package bot

import (
	"fmt"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"time"

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

const (
	Unregistered = "⚡ You are not registered! Press /start to register!"
)
