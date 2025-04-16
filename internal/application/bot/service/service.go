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

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/telebot"
)

type BotService struct {
	srv *http.Server
	bt  *telebot.Bot
}

func New(srv *http.Server, bt *telebot.Bot) (*BotService, error) {
	return &BotService{
		srv: srv,
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
