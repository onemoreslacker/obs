package scrapperservice

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"golang.org/x/sync/errgroup"
)

type Runnable interface {
	Run(ctx context.Context) error
}

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type Updater interface {
	Send(ctx context.Context, chatID int64, url, description string) error
}

type ScrapperService struct {
	fetcher  Runnable
	notifier Runnable
	updater  Updater
	srv      Server
}

func New(
	fetcher,
	notifier Runnable,
	updater Updater,
	srv Server,
) *ScrapperService {
	return &ScrapperService{
		fetcher:  fetcher,
		notifier: notifier,
		updater:  updater,
		srv:      srv,
	}
}

func (s *ScrapperService) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := s.fetcher.Run(ctx); err != nil {
			return fmt.Errorf("scrapper service: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := s.notifier.Run(ctx); err != nil {
			return fmt.Errorf("scrapper service: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		srvErr := make(chan error, 1)
		go func() {
			srvErr <- s.srv.ListenAndServe()
		}()

		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(
				context.Background(),
				config.ShutdownTimeout,
			)
			defer cancel()

			if err := s.srv.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("scrapper service: failed to shutdown server: %w", err)
			}

			if err := <-srvErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("scrapper service: %w", err)
			}

			return nil
		case err := <-srvErr:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("scrapper service: %w", err)
			}
			return nil
		}
	})

	return g.Wait()
}
