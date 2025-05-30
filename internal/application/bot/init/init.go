package botinit

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	botapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/bot"
	botserver "github.com/es-debug/backend-academy-2024-go-template/internal/api/servers/bot"
	botservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/service"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/telebot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func BotCommands(tgc *tgbotapi.BotAPI) error {
	commands := make([]tgbotapi.BotCommand, len(config.Descriptions))
	for i, data := range config.Descriptions {
		commands[i] = tgbotapi.BotCommand{
			Command:     data.Name,
			Description: data.Description,
		}
	}

	commandsConfig := tgbotapi.NewSetMyCommands(commands...)
	if _, err := tgc.Request(commandsConfig); err != nil {
		return fmt.Errorf("failed to load bot commands: %w", err)
	}

	return nil
}

func Config() (*config.Config, error) {
	configFileName := flag.String("config", "", "path to config file")

	flag.Parse()

	cfg, err := config.New(*configFileName)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func TelegramAPI(cfg *config.Config) (*tgbotapi.BotAPI, error) {
	tgc, err := tgbotapi.NewBotAPI(cfg.Secrets.BotToken)
	if err != nil {
		return nil, fmt.Errorf("telegram api was not initialized: %w", err)
	}

	return tgc, nil
}

func ScrapperClient(cfg *config.Config) (sclient.ClientInterface, error) {
	server := url.URL{
		Scheme: config.Scheme,
		Host:   net.JoinHostPort(cfg.Serving.ScrapperHost, cfg.Serving.ScrapperPort),
	}

	client, err := sclient.NewClient(server.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create scrapper client: %w", err)
	}

	return client, nil
}

func BotServer(tgc *tgbotapi.BotAPI, cfg *config.Config) *http.Server {
	api := botapi.New(tgc)
	return botserver.New(cfg.Serving, api)
}

func BotService(srv *http.Server, bt *telebot.Bot) (*botservice.BotService, error) {
	service, err := botservice.New(srv, bt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize bot service: %w", err)
	}

	return service, err
}

func Telebot(client sclient.ClientInterface, tgc *tgbotapi.BotAPI) *telebot.Bot {
	return telebot.New(client, tgc)
}
