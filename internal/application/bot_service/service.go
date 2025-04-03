package botservice

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	botapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/bot_api"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/services/bot"
	scrcl "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/scrapper"
	srvb "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/servers/bot_server"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotService struct {
	srv *http.Server
	bt  *bot.Bot
}

func New(cfg *config.Config) (*BotService, error) {
	tgc, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	commands := make([]tgbotapi.BotCommand, len(config.Descriptions))
	for i, data := range config.Descriptions {
		commands[i] = tgbotapi.BotCommand{
			Command:     data.Name,
			Description: data.Description,
		}
	}

	commandsConfig := tgbotapi.NewSetMyCommands(commands...)

	if _, err := tgc.Request(commandsConfig); err != nil {
		return nil, err
	}

	client, err := scrcl.NewClient("http://" + net.JoinHostPort(cfg.Host, cfg.ScrapperPort))
	if err != nil {
		return nil, err
	}

	bt, err := bot.New(client, tgc, cfg)
	if err != nil {
		return nil, err
	}

	api := botapi.New(tgc)

	return &BotService{
		srv: srvb.New(cfg, api),
		bt:  bt,
	}, nil
}

func (s *BotService) Run() error {
	srvErr := make(chan error)

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	go func() {
		s.bt.Run()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-srvErr:
		slog.Error(
			"server error",
			slog.String("error", err.Error()),
			slog.String("service", "bot"),
		)
	case sig := <-stop:
		slog.Info(
			"received shutdown signal",
			slog.String("signal", sig.String()),
			slog.String("service", "bot"),
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
