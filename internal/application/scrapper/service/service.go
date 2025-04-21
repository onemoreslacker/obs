package scrapperservice

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/notifier"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/updater"
)

type ScrapperService struct {
	upd *updater.Updater
	nt  *notifier.Notifier
	srv *http.Server
}

func New(upd *updater.Updater, nt *notifier.Notifier, srv *http.Server) *ScrapperService {
	return &ScrapperService{
		upd: upd,
		nt:  nt,
		srv: srv,
	}
}

func (s *ScrapperService) Run(ctx context.Context) error {
	srvErr := make(chan error, 1)

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	notifierErr := make(chan error, 1)

	go func() {
		if err := s.nt.Run(ctx); err != nil {
			notifierErr <- err
		}
	}()

	s.upd.Run(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-srvErr:
		slog.Error(
			"server error",
			slog.String("msg", err.Error()),
			slog.String("service", "scrapper"),
		)
	case err := <-notifierErr:
		slog.Error(
			"notifier error",
			slog.String("msg", err.Error()),
			slog.String("service", "scrapper"),
		)
	case sig := <-stop:
		slog.Info(
			"received shutdown signal",
			slog.String("signal", sig.String()),
			slog.String("service", "scrapper"),
		)
	}

	if err := s.srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
