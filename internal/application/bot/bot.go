package bot

import (
	"net"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/commands"
	scrcl "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/scrapper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	tgb            TelegramBotAPI
	scrapperClient scrcl.ClientInterface
	currentCommand commands.Command
	cfg            *config.Config
}

func New(tgbot TelegramBotAPI, cfg *config.Config) (*Bot, error) {
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

func (b *Bot) Run() error {
	updates := b.configureUpdates()

	for update := range updates {
		msg, query := update.Message, update.CallbackQuery

		if msg == nil && query == nil {
			continue
		}

		if query != nil {
			if err := b.QueryHandler(query); err != nil {
				return err
			}
		} else {
			if err := b.MessageHandler(msg.Chat.ID, msg.Text); err != nil {
				return err
			}
		}
	}

	return nil
}

//go:generate mockery --name=TelegramBotAPI --structname=MockTeletramBotAPI --filename=mock_telegram_bot_api_test.go --outpkg=bot_test --output=.
type TelegramBotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
}
