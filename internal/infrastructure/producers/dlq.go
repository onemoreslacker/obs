package producers

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type DLQHandler struct {
	prod Producer
}

func NewDLQHandler(prod Producer) *DLQHandler {
	return &DLQHandler{
		prod: prod,
	}
}

func (d *DLQHandler) Send(ctx context.Context, msg kafka.Message, reason string) error {
	dlqMsg := kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
		Headers: append(msg.Headers, kafka.Header{
			Key:   "dlq-reason",
			Value: []byte(reason),
		}),
	}

	if err := d.prod.WriteMessages(ctx, dlqMsg); err != nil {
		return fmt.Errorf("failed to write message to dlq: %w", err)
	}

	return nil
}
