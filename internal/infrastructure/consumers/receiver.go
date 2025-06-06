package consumers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	ReadMessage(context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type DLQHandler interface {
	Send(ctx context.Context, msg kafka.Message, reason string) error
}

type Sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type Deserializer interface {
	Deserialize(r io.Reader, v any) error
}

type UpdateReceiver struct {
	cons         Consumer
	dlqHandler   DLQHandler
	tc           Sender
	deserializer Deserializer
}

func NewUpdateReceiver(
	cons Consumer,
	dlqHandler DLQHandler,
	tc Sender,
	deserializer Deserializer,
) *UpdateReceiver {
	return &UpdateReceiver{
		cons:         cons,
		dlqHandler:   dlqHandler,
		tc:           tc,
		deserializer: deserializer,
	}
}

func (u *UpdateReceiver) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}

		if err := u.ProcessMessage(ctx); err != nil {
			slog.Error(
				"update receiver: failed to process message",
				slog.String("msg", err.Error()),
			)
		}
	}
}

func (u *UpdateReceiver) ProcessMessage(ctx context.Context) error {
	msg, err := u.cons.ReadMessage(ctx)
	if err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}

	if err = u.HandleUpdate(msg); err != nil {
		if err = u.dlqHandler.Send(ctx, msg, err.Error()); err != nil {
			return fmt.Errorf("failed to write message to dlq: %w", err)
		}
	}

	if err = u.cons.CommitMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to commit message: %w", err)
	}

	return nil
}

func (u *UpdateReceiver) HandleUpdate(msg kafka.Message) error {
	var update models.KafkaUpdate
	if err := u.deserializer.Deserialize(bytes.NewReader(msg.Value), &update); err != nil {
		return fmt.Errorf("failed to deserialize update: %w", err)
	}

	if _, err := u.tc.Send(tgbotapi.NewMessage(
		update.ChatID,
		fmt.Sprintf(`✨ New update via %s!

%s`, update.Url, update.Description),
	)); err != nil {
		return fmt.Errorf("failed to sent update to telegram: %w", err)
	}

	return nil
}

func (u *UpdateReceiver) Stop(_ context.Context) error {
	if err := u.cons.Close(); err != nil {
		return fmt.Errorf("update receiver: failed to close consumer: %w", err)
	}

	return nil
}
