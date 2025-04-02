package scrapperservice

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/services/scrapper"
	scrapperserver "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/servers/scrapper_server"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/storage"
)

type ScrapperService struct {
	scr *scrapper.Scrapper
	srv *http.Server
}

func New(cfg *config.Config) (*ScrapperService, error) {
	links := storage.NewLinksRepository()

	scr, err := scrapper.New(cfg, links)
	if err != nil {
		return nil, err
	}

	api := scrapperapi.New(links)

	return &ScrapperService{
		scr: scr,
		srv: scrapperserver.New(cfg, api),
	}, nil
}

func (s *ScrapperService) Run() error {
	srvErr := make(chan error, 1)

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	scrapperErr := make(chan error, 1)

	go func() {
		if err := s.scr.Run(); err != nil {
			scrapperErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-srvErr:
		slog.Error(
			"server error",
			slog.String("msg", err.Error()),
			slog.String("service", "scrapper"),
		)
	case err := <-scrapperErr:
		slog.Error(
			"scrapper error",
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
