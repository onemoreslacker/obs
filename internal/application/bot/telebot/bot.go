package telebot

import (
	"context"
	"log/slog"
	"net/http"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}

type ScrapperClient interface {
	DeleteLinks(ctx context.Context, params *sclient.DeleteLinksParams, body sclient.DeleteLinksJSONRequestBody,
		reqEditors ...sclient.RequestEditorFn) (*http.Response, error)
	GetLinks(ctx context.Context, params *sclient.GetLinksParams,
		reqEditors ...sclient.RequestEditorFn) (*http.Response, error)
	PostLinks(ctx context.Context, params *sclient.PostLinksParams, body sclient.PostLinksJSONRequestBody,
		reqEditors ...sclient.RequestEditorFn) (*http.Response, error)
	DeleteTgChatId(ctx context.Context, id int64, reqEditors ...sclient.RequestEditorFn) (*http.Response, error)
	GetTgChatId(ctx context.Context, id int64, reqEditors ...sclient.RequestEditorFn) (*http.Response, error)
	PostTgChatId(ctx context.Context, id int64, reqEditors ...sclient.RequestEditorFn) (*http.Response, error)
}

type Command interface {
	Stage() (string, bool)
	Validate(input string) error
	Done() bool
	Request() string
	Name() string
}

type Bot struct {
	Tgb            TgAPI
	Client         ScrapperClient
	CurrentCommand Command
}

func New(client ScrapperClient, api TgAPI) *Bot {
	return &Bot{
		Tgb:    api,
		Client: client,
	}
}

func (b *Bot) Run(ctx context.Context) {
	updates := b.ConfigureUpdates()

	for update := range updates {
		msg, query := update.Message, update.CallbackQuery

		if msg == nil && query == nil {
			continue
		}

		var reply tgbotapi.MessageConfig

		if query != nil {
			reply = b.QueryHandler(query)
		} else {
			reply = b.MessageHandler(msg)
		}

		if _, err := b.Tgb.Send(reply); err != nil {
			slog.Error(
				"telebot: failed to reply",
				slog.String("msg", err.Error()),
				slog.String("reply", reply.Text),
				slog.String("service", "bot"),
			)
		}
	}
}

const (
	Start   = "start"
	Help    = "help"
	Cancel  = "cancel"
	Track   = "track"
	Untrack = "untrack"
	List    = "list"
)
