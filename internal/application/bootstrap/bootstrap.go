package bootstrap

import (
	"flag"
	"log/slog"
	"net"
	"net/http"
	"net/url"

	botapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/bot_api"
	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
	botservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/service"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/telebot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/core"
	scrapperservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/service"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	botclient "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/bot"
	scrcl "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/scrapper"
	botserver "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/servers/bot"
	scrapperserver "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/storage"
	"github.com/go-co-op/gocron/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func LoadConfig() (*config.Config, error) {
	configFileName := flag.String("config", "config/config.yaml", "path to config file")

	flag.Parse()

	cfg, err := config.Load(*configFileName)
	if err != nil {
		slog.Error(
			"Config was not loaded",
			slog.String("msg", err.Error()),
		)

		return nil, err
	}

	return cfg, nil
}

func InitPool(cfg *config.Config) (*pgxpool.Pool, error) {
	if cfg.Database.AccessType == "in-memory" {
		return nil, nil
	}

	pool, err := storage.NewPool(cfg)
	if err != nil {
		slog.Error("Pool was not created", slog.String("msg", err.Error()))
		return nil, err
	}

	return pool, nil
}

func InitRepository(cfg *config.Config, pool *pgxpool.Pool) (storage.LinksRepository, error) {
	repository, err := storage.New(cfg, pool)
	if err != nil {
		slog.Error(
			"Repository was not initialized",
			slog.String("msg", err.Error()),
		)

		return nil, err
	}

	return repository, nil
}

func InitBotClient(cfg *config.Config) (*botclient.Client, error) {
	server := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(cfg.Serving.BotHost, cfg.Serving.BotPort),
	}

	client, err := botclient.NewClient(server.String())
	if err != nil {
		slog.Error("Failed to create bot client", slog.String("msg", err.Error()))
		return nil, err
	}

	return client, nil
}

func InitScheduler() (gocron.Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		slog.Error("Failed to create scheduler", slog.String("msg", err.Error()))
		return nil, err
	}

	return scheduler, nil
}

func InitScrapper(client *botclient.Client, repo storage.LinksRepository, scheduler gocron.Scheduler) (*core.Scrapper, error) {
	scr, err := core.New(client, repo, scheduler)
	if err != nil {
		slog.Error("Failed to create scrapper", slog.String("msg", err.Error()))
		return nil, err
	}

	return scr, nil
}

func InitScrapperServer(cfg *config.Config, repo storage.LinksRepository) *http.Server {
	api := scrapperapi.New(repo)
	return scrapperserver.New(cfg, api)
}

func InitScrapperService(scr *core.Scrapper, srv *http.Server) (*scrapperservice.ScrapperService, error) {
	service, err := scrapperservice.New(scr, srv)
	if err != nil {
		slog.Error("Failed to initialize scrapper service", slog.String("msg", err.Error()))
		return nil, err
	}

	return service, nil
}

func InitTelegramAPI(cfg *config.Config) (*tgbotapi.BotAPI, error) {
	tgc, err := tgbotapi.NewBotAPI(cfg.Secrets.BotToken)
	if err != nil {
		slog.Error("Telegram API was not initialized",
			slog.String("msg", err.Error()),
		)

		return nil, err
	}

	return tgc, nil
}

func InitBotCommands(tgc *tgbotapi.BotAPI) error {
	commands := make([]tgbotapi.BotCommand, len(config.Descriptions))
	for i, data := range config.Descriptions {
		commands[i] = tgbotapi.BotCommand{
			Command:     data.Name,
			Description: data.Description,
		}
	}

	commandsConfig := tgbotapi.NewSetMyCommands(commands...)
	if _, err := tgc.Request(commandsConfig); err != nil {
		slog.Error("Failed to load bot commands",
			slog.String("msg", err.Error()),
		)

		return err
	}

	if _, err := tgc.Request(commandsConfig); err != nil {
		return err
	}

	return nil
}

func InitScrapperClient(cfg *config.Config) (scrcl.ClientInterface, error) {
	client, err := scrcl.NewClient("http://" + net.JoinHostPort(cfg.Serving.ScrapperHost, cfg.Serving.ScrapperPort))
	if err != nil {
		slog.Error(
			"Failed to create scrapper client",
			slog.String("msg", err.Error()),
		)

		return nil, err
	}

	return client, nil
}

func InitTelebot(client scrcl.ClientInterface, tgc *tgbotapi.BotAPI) (*telebot.Bot, error) {
	bt, err := telebot.New(client, tgc)
	if err != nil {
		slog.Error(
			"Failed to initialize telebot",
			slog.String("msg", err.Error()),
		)

		return nil, err
	}

	return bt, nil
}

func InitBotServer(tgc *tgbotapi.BotAPI, cfg *config.Config) *http.Server {
	api := botapi.New(tgc)
	return botserver.New(cfg, api)
}

func InitBotService(srv *http.Server, bt *telebot.Bot) (*botservice.BotService, error) {
	service, err := botservice.New(srv, bt)
	if err != nil {
		slog.Error(
			"Failed to initialize bot service",
			slog.String("msg", err.Error()),
		)

		return nil, err
	}

	return service, err
}
