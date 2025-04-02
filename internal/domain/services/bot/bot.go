package bot

import (
	"log/slog"
	"net"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	scrcl "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/scrapper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	tgb            TgAPI
	scrapperClient scrcl.ClientInterface
	currentCommand Command
	cfg            *config.Config
}

//go:generate mockery --name TgAPI --structname MockTgAPI --filename mock_tg_api_test.go --outpkg bot_test --output .
type TgAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}

type Command interface {
	Stage() (string, bool)
	Validate(input string) error
	Done() bool
	Request() string
	Name() string
}

func New(tgbot TgAPI, cfg *config.Config) (*Bot, error) {
	client, err := scrcl.NewClient("http://" + net.JoinHostPort(cfg.Host, cfg.ScrapperPort))
	if err != nil {
		return nil, err
	}

	return &Bot{
		tgb:            tgbot,
		scrapperClient: client,
		cfg:            cfg,
	}, nil
}

func (b *Bot) Run() {
	updates := b.configureUpdates()

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

		if _, err := b.tgb.Send(reply); err != nil {
			slog.Error(
				"failed to reply",
				slog.String("msg", err.Error()),
				slog.String("reply", reply.Text),
				slog.String("service", "bot"),
			)
		}
	}
}

const (
	start   = "start"
	help    = "help"
	cancel  = "cancel"
	track   = "track"
	untrack = "untrack"
	list    = "list"
)
