package fetcher

import (
	"context"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/svcident"
)

type Option func(*updaterConfig)

func WithCustomBatchSize(batchSize uint64) Option {
	return func(cfg *updaterConfig) {
		cfg.batchSize = batchSize
	}
}

func WithCustomWorkersNumber(workersNum int) Option {
	return func(cfg *updaterConfig) {
		cfg.workersNum = workersNum
	}
}

type updaterConfig struct {
	batchSize  uint64
	workersNum int
}

func (f *Fetcher) CheckActivity(ctx context.Context, link string) (bool, error) {
	var (
		updates []models.Update
		err     error
	)

	service, err := svcident.FromLink(link)
	if err != nil {
		return false, err
	}

	switch service {
	case config.GitHub:
		updates, err = f.GitHub.RetrieveUpdates(ctx, link)
	case config.StackOverflow:
		updates, err = f.Stack.RetrieveUpdates(ctx, link)
	}

	if err != nil {
		return false, err
	}

	if len(updates) == 0 {
		return false, nil
	}

	createdAt, err := time.Parse(time.RFC3339, updates[0].CreatedAt)
	if err != nil {
		return false, err
	}

	return createdAt.After(getCutoff()), nil
}

func getCutoff() time.Time {
	yesterday := time.Now().AddDate(0, 0, -1)
	cutoff := time.Date(
		yesterday.Year(),
		yesterday.Month(),
		yesterday.Day(),
		10,
		0,
		0,
		0,
		yesterday.Location(),
	)

	return cutoff
}
