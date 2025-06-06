package scrapperservice

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

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

type UpdateSender interface {
	Send(ctx context.Context, chatID int64, url, description string) error
	Stop() error
}

type ScrapperService struct {
	updater  Runnable
	notifier Runnable
	sender   UpdateSender
	srv      Server
}

func New(
	updater, notifier Runnable,
	sender UpdateSender,
	srv Server,
) *ScrapperService {
	return &ScrapperService{
		updater:  updater,
		notifier: notifier,
		sender:   sender,
		srv:      srv,
	}
}

func (s *ScrapperService) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := s.updater.Run(ctx); err != nil {
			return fmt.Errorf("updater error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := s.notifier.Run(ctx); err != nil {
			return fmt.Errorf("notifier error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	runErr := g.Wait()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	errs := []error{runErr}
	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		errs = append(errs, fmt.Errorf("failed to shutdown server: %w", err))
	}

	if err := s.sender.Stop(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close update sender: %w", err))
	}

	return errors.Join(errs...)
}
