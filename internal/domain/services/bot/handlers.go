package bot

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) commandRequest(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	slog.Info(
		"bot requests scrapper",
		slog.String("cmd", b.currentCommand.Name()),
		slog.String("service", "bot"),
	)

	defer func() { b.currentCommand = nil }()

	reply := b.currentCommand.Request()

	return tgbotapi.NewMessage(msg.Chat.ID, reply)
}

func (b *Bot) QueryHandler(query *tgbotapi.CallbackQuery) tgbotapi.MessageConfig {
	msg, input := query.Message, query.Data

	slog.Info(
		"query content",
		slog.String("msg", msg.Text),
		slog.String("data", input),
		slog.String("service", "bot"),
	)

	callback := tgbotapi.NewCallback(query.ID, input)
	if _, err := b.tgb.Request(callback); err != nil {
		return tgbotapi.NewMessage(msg.Chat.ID, FailedConfirmation)
	}

	if b.currentCommand == nil {
		return tgbotapi.NewMessage(msg.Chat.ID, UndefinedConfirmation)
	}

	msg.Text = input

	return b.handleState(msg)
}

func (b *Bot) MessageHandler(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var response tgbotapi.MessageConfig

	slog.Info(
		"current command",
		slog.String("cmd", msg.Command()),
	)

	switch msg.Command() {
	case start:
		response = b.handleStart(msg)
	case help:
		response = b.handleHelp(msg)
	case cancel:
		response = b.handleCancel(msg)
	default:
		response = b.handleState(msg)
	}

	return response
}

func (b *Bot) handleState(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	if err := b.isRegistered(msg.Chat.ID); err != nil {
		return tgbotapi.NewMessage(
			msg.Chat.ID,
			NeedRegistration,
		)
	}

	if b.currentCommand == nil {
		return b.InitializeCommand(msg)
	}

	if err := b.currentCommand.Validate(msg.Text); err != nil {
		return b.configureReply(msg)
	}

	if b.currentCommand.Done() {
		return b.commandRequest(msg)
	}

	return b.configureReply(msg)
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(
		msg.Chat.ID, b.withAuthorization(
			msg.Chat.ID,
			constructHelpMessage,
		),
	)
}

func (b *Bot) handleStart(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	resp, err := b.scrapperClient.PostTgChatId(context.Background(), msg.Chat.ID)
	if err != nil {
		slog.Error(
			"registration failed",
			slog.String("msg", err.Error()),
		)

		if !errors.Is(err, scrapperapi.ErrChatAlreadyExists) {
			return tgbotapi.NewMessage(msg.Chat.ID, FailedRegistration)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return tgbotapi.NewMessage(msg.Chat.ID, RepeatedRegistration)
	}

	return tgbotapi.NewMessage(msg.Chat.ID, SuccessfulRegistration)
}

func (b *Bot) handleCancel(msg *tgbotapi.Message) tgbotapi.MessageConfig {
	if b.currentCommand == nil {
		return tgbotapi.NewMessage(msg.Chat.ID, UndefinedCancel)
	}

	return tgbotapi.NewMessage(msg.Chat.ID, SuccessfulCancel)
}

const (
	FailedRegistration = "💥 Failed to register!"
	FailedConfirmation = "💥 Nothing to confirm!"

	NeedRegistration      = "⚡️ Please, register with /start command!"
	RepeatedRegistration  = "⚡️ You are already registered!"
	UndefinedConfirmation = "⚡️ Nothing to confirm!"
	UndefinedCancel       = "⚡️ Nothing to cancel!"

	SuccessfulCancel       = "↩️ Command cancelled!"
	SuccessfulRegistration = "✨ You are successfully registered! Use /help to get a list of available commands."
)
