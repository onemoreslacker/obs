package telebot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) WithRegistration(handler func(ctx context.Context, msg *tgbotapi.Message) tgbotapi.MessageConfig,
) func(ctx context.Context, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	return func(ctx context.Context, msg *tgbotapi.Message) tgbotapi.MessageConfig {
		if err := b.IsRegistered(msg.Chat.ID); err != nil {
			return tgbotapi.NewMessage(
				msg.Chat.ID,
				"⚡️ Please, register with /start command!",
			)
		}

		return handler(ctx, msg)
	}
}
