package telebot

import (
	"context"
	"log/slog"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) CommandRequest(ctx context.Context, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	defer func() { b.CommandStates.Store(msg.Chat.ID, nil) }()

	currentCommand, _ := b.CommandStates.Load(msg.Chat.ID)

	reply, err := currentCommand.Request(ctx)
	if err != nil {
		slog.Error(
			"telebot: current command request failed",
			slog.String("msg", err.Error()),
		)
	}

	return tgbotapi.NewMessage(msg.Chat.ID, reply)
}

func (b *Bot) QueryHandler(ctx context.Context, query *tgbotapi.CallbackQuery) tgbotapi.MessageConfig {
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

	if _, exists := b.CommandStates.Load(msg.Chat.ID); !exists {
		return tgbotapi.NewMessage(msg.Chat.ID, "⚡️ Nothing to confirm!")
	}

	msg.Text = input

	return b.HandleState(ctx, msg)
}

func (b *Bot) MessageHandler(ctx context.Context, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	var response tgbotapi.MessageConfig

	slog.Info(
		"telebot: current command",
		slog.String("cmd", msg.Command()),
		slog.Int64("chatID", msg.Chat.ID),
	)

	switch msg.Command() {
	case Start:
		response = b.HandleStart(ctx, msg)
	case Help:
		response = b.WithRegistration(b.HandleHelp)(ctx, msg)
	case Cancel:
		response = b.WithRegistration(b.HandleCancel)(ctx, msg)
	default:
		response = b.WithRegistration(b.HandleState)(ctx, msg)
	}

	return response
}

func (b *Bot) HandleState(ctx context.Context, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	currentCommand, exists := b.CommandStates.Load(msg.Chat.ID)
	if !exists || currentCommand == nil {
		return b.InitializeCommand(msg)
	}

	if err := currentCommand.Validate(msg.Text); err != nil {
		return b.ConfigureReply(msg)
	}

	if currentCommand.Done() {
		return b.CommandRequest(ctx, msg)
	}

	return b.ConfigureReply(msg)
}

func (b *Bot) HandleHelp(_ context.Context, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(msg.Chat.ID, ConstructHelpMessage())
}

func (b *Bot) HandleStart(ctx context.Context, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	resp, err := b.Client.PostTgChatId(ctx, msg.Chat.ID)
	if err != nil {
		slog.Error(
			"telebot: failed to handle /start",
			slog.String("msg", err.Error()),
		)

		return tgbotapi.NewMessage(msg.Chat.ID, "💥 Failed to register!")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return tgbotapi.NewMessage(msg.Chat.ID, "⚡️ You are already registered!")
	}

	return tgbotapi.NewMessage(msg.Chat.ID,
		"✨ You are successfully registered! Use /help to get a list of available commands.")
}

func (b *Bot) HandleCancel(_ context.Context, msg *tgbotapi.Message) tgbotapi.MessageConfig {
	currentCommand, exists := b.CommandStates.Load(msg.Chat.ID)
	if !exists || currentCommand == nil {
		return tgbotapi.NewMessage(msg.Chat.ID, "⚡️ Nothing to cancel!")
	}

	b.CommandStates.Store(msg.Chat.ID, nil)

	return tgbotapi.NewMessage(msg.Chat.ID, "↩️ Command cancelled!")
}
