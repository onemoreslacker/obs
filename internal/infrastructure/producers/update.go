package producers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/service"
	"github.com/segmentio/kafka-go"
)

type Serializer interface {
	Serialize(v any) ([]byte, error)
}

type UpdatePublisher struct {
	producer   service.Producer
	serializer Serializer
}

func NewUpdatePublisher(producer service.Producer, serializer Serializer) *UpdatePublisher {
	return &UpdatePublisher{
		producer:   producer,
		serializer: serializer,
	}
}

func (u *UpdatePublisher) Send(ctx context.Context, chatID int64, url, description string) error {
	data, err := u.serializer.Serialize(models.KafkaUpdate{
		ChatID:      chatID,
		Url:         url,
		Description: description,
	})
	if err != nil {
		return fmt.Errorf("update publisher: failed to serialize update: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(strconv.FormatInt(chatID, 10)),
		Value: data,
	}

	if err = u.producer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("update publisher: failed to write message: %w", err)
	}

	return nil
}
