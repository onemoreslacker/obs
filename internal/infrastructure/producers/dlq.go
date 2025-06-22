package producers

import (
	"context"
	"fmt"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/service"
	"github.com/segmentio/kafka-go"
)

type DLQPublisher struct {
	producer service.Producer
}

func NewDLQPublisher(producer service.Producer) *DLQPublisher {
	return &DLQPublisher{
		producer: producer,
	}
}

func (d *DLQPublisher) Send(ctx context.Context, msg kafka.Message, reason string) error {
	dlqMsg := kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
		Headers: append(msg.Headers, kafka.Header{
			Key:   "dlq-reason",
			Value: []byte(reason),
		}),
	}

	if err := d.producer.WriteMessages(ctx, dlqMsg); err != nil {
		return fmt.Errorf("dlq publisher: failed to write message: %w", err)
	}

	return nil
}
