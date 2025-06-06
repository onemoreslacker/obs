package botservice

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"golang.org/x/sync/errgroup"
)

type UpdateReceiver interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Runnable interface {
	Run(ctx context.Context) error
}

type BotService struct {
	receiver UpdateReceiver
	bot      Runnable
}

func New(receiver UpdateReceiver, bot Runnable) *BotService {
	return &BotService{
		receiver: receiver,
		bot:      bot,
	}
}

func (s *BotService) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := s.receiver.Start(ctx); err != nil {
			return fmt.Errorf("update receiver error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := s.bot.Run(ctx); err != nil {
			return fmt.Errorf("telebot error: %w", err)
		}
		return nil
	})

	runErr := g.Wait()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	errs := []error{runErr}
	if err := s.receiver.Stop(shutdownCtx); err != nil {
		errs = append(errs, fmt.Errorf("failed to stop update receiver: %w", err))
	}

	return errors.Join(errs...)
}
