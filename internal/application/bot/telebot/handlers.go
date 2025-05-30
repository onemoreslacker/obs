package telebot

import (
	"context"
	"log/slog"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) CommandRequest(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	defer func() { b.CurrentCommand = nil }()

	reply := b.CurrentCommand.Request()

	return tgbotapi.NewMessage(msg.Chat.ID, reply)
}

func (b *Bot) QueryHandler(query *tgbotapi.CallbackQuery) tgbotapi.MessageConfig {
	msg, input := query.Message, query.Data

	slog.Info(
		"telebot: query content",
		slog.String("msg", msg.Text),
		slog.String("data", input),
		slog.String("service", "bot"),
	)

	callback := tgbotapi.NewCallback(query.ID, input)
	resp, err := b.Tgb.Request(callback)
	if err != nil {
		return tgbotapi.NewMessage(msg.Chat.ID, "💥 Telegram API request failed!")
	}

	if !resp.Ok {
		return tgbotapi.NewMessage(msg.Chat.ID, "💥 Failed to confirm!")
	}

	if b.CurrentCommand == nil {
		return tgbotapi.NewMessage(msg.Chat.ID, "⚡️ Nothing to confirm!")
	}

	msg.Text = input

	return b.HandleState(msg)
}

func (b *Bot) MessageHandler(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var response tgbotapi.MessageConfig

	slog.Info(
		"telebot: current command",
		slog.String("cmd", msg.Command()),
	)

	switch msg.Command() {
	case Start:
		response = b.HandleStart(msg)
	case Help:
		response = b.WithRegistration(b.HandleHelp)(msg)
	case Cancel:
		response = b.WithRegistration(b.HandleCancel)(msg)
	default:
		response = b.WithRegistration(b.HandleState)(msg)
	}

	return response
}

func (b *Bot) HandleState(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	if b.CurrentCommand == nil {
		return b.InitializeCommand(msg)
	}

	if err := b.CurrentCommand.Validate(msg.Text); err != nil {
		return b.ConfigureReply(msg)
	}

	if b.CurrentCommand.Done() {
		return b.CommandRequest(msg)
	}

	return b.ConfigureReply(msg)
}

func (b *Bot) HandleHelp(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(msg.Chat.ID, ConstructHelpMessage())
}

func (b *Bot) HandleStart(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	resp, err := b.Client.PostTgChatId(context.Background(), msg.Chat.ID)
	if err != nil {
		return tgbotapi.NewMessage(msg.Chat.ID, "💥 Failed to register!")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tgbotapi.NewMessage(msg.Chat.ID, "⚡️ You are already registered!")
	}

	return tgbotapi.NewMessage(msg.Chat.ID,
		"✨ You are successfully registered! Use /help to get a list of available commands.")
}

func (b *Bot) HandleCancel(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	if b.CurrentCommand == nil {
		return tgbotapi.NewMessage(msg.Chat.ID, "⚡️ Nothing to cancel!")
	}

	b.CurrentCommand = nil

	return tgbotapi.NewMessage(msg.Chat.ID, "↩️ Command cancelled!")
}
