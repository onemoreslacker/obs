package sender

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/config"
)

type UpdateSender interface {
	Send(ctx context.Context, chatID int64, url, description string) error
	Stop() error
}

func New(
	syncSender UpdateSender,
	asyncSender UpdateSender,
	cfg *config.Transport,
) (UpdateSender, error) {
	switch cfg.Mode {
	case config.Sync:
		return syncSender, nil
	case config.Async:
		return asyncSender, nil
	}

	return nil, ErrUnknownTransportMode
}
