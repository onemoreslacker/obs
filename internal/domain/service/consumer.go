package service

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	ReadMessage(context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, messages ...kafka.Message) error
}
