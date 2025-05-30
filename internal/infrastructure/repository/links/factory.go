package links

import (
	"context"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Add(ctx context.Context, url string) (int64, error)
	Delete(ctx context.Context, linkID int64) error
	Touch(ctx context.Context, linkID int64) error
	UpdateActivity(ctx context.Context, linkID int64, status bool) error
	GetBatch(ctx context.Context, batchSize uint64) ([]sapi.LinkResponse, error)
}

type Option func(Repository)

func WithTimeProvider(provider func() time.Time) Option {
	return func(r Repository) {
		switch repo := r.(type) {
		case *SquirrelRepository:
			repo.now = provider
		case *SQLRepository:
			repo.now = provider
		}
	}
}

func New(cfg config.Database, pool *pgxpool.Pool) Repository {
	switch cfg.Access {
	case config.Orm:
		return NewSquirrelRepository(pool)
	case config.Sql:
		return NewSQLRepository(pool)
	}

	return nil
}
