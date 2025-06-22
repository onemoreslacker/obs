package botservice

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
)

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type UpdateSubscriber interface {
	Start(
		ctx context.Context,
		process func(ctx context.Context, msg kafka.Message) error,
	) error
}

type Processor interface {
	Process(ctx context.Context, msg kafka.Message) error
}

type Runnable interface {
	Run(ctx context.Context) error
}

type BotService struct {
	srv              Server
	updateSubscriber UpdateSubscriber
	processor        Processor
	bot              Runnable
}

func New(
	srv Server,
	updateSubscriber UpdateSubscriber,
	processor Processor,
	bot Runnable,
) *BotService {
	return &BotService{
		srv:              srv,
		updateSubscriber: updateSubscriber,
		processor:        processor,
		bot:              bot,
	}
}

func (b *BotService) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := b.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("bot service: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := b.updateSubscriber.Start(ctx, b.processor.Process); err != nil {
			return fmt.Errorf("bot service: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := b.bot.Run(ctx); err != nil {
			return fmt.Errorf("bot service: %w", err)
		}
		return nil
	})

	runErr := g.Wait()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	errs := []error{runErr}
	if err := b.srv.Shutdown(shutdownCtx); err != nil {
		errs = append(errs, fmt.Errorf("failed to stop bot server: %w", err))
	}

	return errors.Join(errs...)
}
