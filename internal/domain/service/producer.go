package service

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	WriteMessages(ctx context.Context, messages ...kafka.Message) error
}
