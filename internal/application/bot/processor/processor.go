package processor

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/segmentio/kafka-go"
)

type TelegramSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type Deserializer interface {
	Deserialize(r io.Reader, v any) error
}

type DLQSender interface {
	Send(ctx context.Context, msg kafka.Message, reason string) error
}

type Processor struct {
	telegramSender TelegramSender
	deserializer   Deserializer
	dlqSender      DLQSender
}

func New(
	telegramSender TelegramSender,
	deserializer Deserializer,
	dlqSender DLQSender,
) *Processor {
	return &Processor{
		telegramSender: telegramSender,
		deserializer:   deserializer,
		dlqSender:      dlqSender,
	}
}

func (p *Processor) Process(ctx context.Context, msg kafka.Message) error {
	var update models.KafkaUpdate
	if err := p.deserializer.Deserialize(bytes.NewReader(msg.Value), &update); err != nil {
		if err = p.dlqSender.Send(ctx, msg,
			fmt.Sprintf("deserialization failed: %s", err.Error())); err != nil {
			return fmt.Errorf("processor: %w", err)
		}
		return fmt.Errorf("processor: failed to deserialize update: %w", err)
	}

	if _, err := p.telegramSender.Send(tgbotapi.NewMessage(
		update.ChatID, fmt.Sprintf(`✨ New update via %s!

%s`, update.Url, update.Description),
	)); err != nil {
		if err = p.dlqSender.Send(ctx, msg,
			fmt.Sprintf("telegram send failed: %s", err.Error())); err != nil {
			return fmt.Errorf("processor: %w", err)
		}
		return fmt.Errorf("processor: failed to send telegram message: %w", err)
	}

	return nil
}
