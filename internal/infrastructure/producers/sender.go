package producers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/segmentio/kafka-go"
)

type Producer interface {
	WriteMessages(context.Context, ...kafka.Message) error
	Close() error
}

type Serializer interface {
	Serialize(v any) ([]byte, error)
}

type UpdateSender struct {
	prod       Producer
	serializer Serializer
}

func NewUpdateSender(prod Producer, serializer Serializer) *UpdateSender {
	return &UpdateSender{
		prod:       prod,
		serializer: serializer,
	}
}

func (u *UpdateSender) Send(ctx context.Context, chatID int64, url, description string) error {
	data, err := u.serializer.Serialize(models.KafkaUpdate{
		ChatID:      chatID,
		Url:         url,
		Description: description,
	})
	if err != nil {
		return fmt.Errorf("update sender: failed to serialize update: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(strconv.FormatInt(chatID, 10)),
		Value: data,
	}

	if err = u.prod.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("update sender: failed to write message: %w", err)
	}

	return nil
}

func (u *UpdateSender) Stop() error {
	if err := u.prod.Close(); err != nil {
		return fmt.Errorf("update sender: failed to close producer: %w", err)
	}

	return nil
}
