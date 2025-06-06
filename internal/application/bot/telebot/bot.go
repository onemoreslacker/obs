package telebot

import (
	"context"
	"log/slog"
	"net/http"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/puzpuzpuz/xsync/v4"
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
	Request(ctx context.Context) (string, error)
	Name() string
}

type Cache interface {
	Add(ctx context.Context, chatID int64, link sclient.LinkResponse) error
	Delete(ctx context.Context, chatID int64, req sclient.RemoveLinkRequest) error
	Get(ctx context.Context, chatID int64) (sclient.ListLinksResponse, error)
}

type Bot struct {
	Tgb           TgAPI
	Client        ScrapperClient
	CommandStates *xsync.Map[int64, Command]
	Cache         Cache
}

func New(client ScrapperClient, api TgAPI, cache Cache) *Bot {
	return &Bot{
		Tgb:           api,
		Client:        client,
		CommandStates: xsync.NewMap[int64, Command](),
		Cache:         cache,
	}
}

func (b *Bot) Run(ctx context.Context) error {
	updates := b.ConfigureUpdates()

	for update := range updates {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}

		msg, query := update.Message, update.CallbackQuery

		if msg == nil && query == nil {
			continue
		}

		go func() {
			var reply tgbotapi.MessageConfig

			if query != nil {
				reply = b.QueryHandler(ctx, query)
			} else {
				reply = b.MessageHandler(ctx, msg)
			}

			if _, err := b.Tgb.Send(reply); err != nil {
				slog.Error(
					"telebot: failed to reply",
					slog.String("msg", err.Error()),
					slog.String("reply", reply.Text),
					slog.String("service", "bot"),
				)
			}
		}()
	}

	return nil
}

const (
	Start   = "start"
	Help    = "help"
	Cancel  = "cancel"
	Track   = "track"
	Untrack = "untrack"
	List    = "list"
)
