package consumers

import (
	"context"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/service"
	"github.com/segmentio/kafka-go"
)

type UpdateSubscriber struct {
	consumer service.Consumer
}

func NewUpdateSubscriber(consumer service.Consumer) *UpdateSubscriber {
	return &UpdateSubscriber{
		consumer: consumer,
	}
}

func (u *UpdateSubscriber) Start(
	ctx context.Context,
	process func(ctx context.Context, msg kafka.Message) error,
) error {
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}

		msg, err := u.consumer.ReadMessage(ctx)
		if err != nil {
			slog.Warn(
				"update subscriber: failed to handle message",
				slog.String("err", err.Error()),
			)
		}

		if err = process(ctx, msg); err != nil {
			slog.Warn(
				"update subscriber: failed to process message",
				slog.String("err", err.Error()),
			)
		}

		if err = u.consumer.CommitMessages(ctx, msg); err != nil {
			slog.Warn(
				"kafka subscriber: failed to commit message",
				slog.String("err", err.Error()),
			)
		}
	}
}
