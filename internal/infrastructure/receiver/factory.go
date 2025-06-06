package receiver

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/config"
)

type UpdateReceiver interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func New(
	syncReceiver UpdateReceiver,
	asyncReceiver UpdateReceiver,
	cfg *config.Transport,
) (UpdateReceiver, error) {
	switch cfg.Mode {
	case config.Sync:
		return syncReceiver, nil
	case config.Async:
		return asyncReceiver, nil
	}

	return nil, ErrUnknownTransportMode
}
