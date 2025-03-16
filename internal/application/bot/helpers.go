package bot

import (
	"bytes"
	"fmt"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"

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
func (b *Bot) constructHelpMessage() string {
	var msg string
	for _, cmd := range b.cfg.Meta.Descriptions {
		msg += fmt.Sprintf("%s - %s\n", cmd.Name, cmd.Description)
	}

	return msg
}

func ConstructListMessage(links []entities.Link) string {
	var buf bytes.Buffer

	for i, link := range links {
		fmt.Fprintf(&buf, "%d. %s\n", i+1, *link.Url)
	}

	return buf.String()
}

func (b *Bot) configureReply(chatID int64) tgbotapi.MessageConfig {
	content, keyboard := b.currentCommand.Stage()

	reply := tgbotapi.NewMessage(chatID, content)
	if keyboard {
		reply.ReplyMarkup = inlineKeyboard
	}

	return reply
}

func (b *Bot) withAuthorization(id int64, next func() string) string {
	if err := b.isRegistered(id); err != nil {
		return b.cfg.Meta.Fails.Unauthorized
	}

	return next()
}

var inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Yes", "Yes"),
		tgbotapi.NewInlineKeyboardButtonData("No", "No"),
	),
)
