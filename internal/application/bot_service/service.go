package botservice

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	botapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/bot_api"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
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

	commands := make([]tgbotapi.BotCommand, len(cfg.Meta.Descriptions))
	for i, data := range cfg.Meta.Descriptions {
		commands[i] = tgbotapi.BotCommand{
			Command:     data.Name,
			Description: data.Description,
		}
	}

	commandsConfig := tgbotapi.NewSetMyCommands(commands...)

	if _, err := tgc.Request(commandsConfig); err != nil {
		return nil, err
	}

	bt, err := bot.New(tgc, cfg)
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
	srvErr := make(chan error, 1)
	botErr := make(chan error, 1)

	go func() {
		if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	go func() {
		if err := s.bt.Run(); err != nil {
			botErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-srvErr:
		slog.Info(
			"Server error",
			"error", err.Error(),
		)
	case err := <-botErr:
		slog.Info(
			"Bot error",
			"error", err.Error(),
		)
	case sig := <-stop:
		slog.Info(
			"Received shutdown signal",
			"signal", sig,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
