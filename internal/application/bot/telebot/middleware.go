package telebot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (b *Bot) WithRegistration(handler func(msg *tgbotapi.Message) tgbotapi.MessageConfig,
) func(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	return func(msg *tgbotapi.Message) tgbotapi.MessageConfig {
		if err := b.IsRegistered(msg.Chat.ID); err != nil {
			return tgbotapi.NewMessage(
				msg.Chat.ID,
				"⚡️ Please, register with /start command!",
			)
		}

		return handler(msg)
	}
}
