package tags

import (
	"context"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Add(ctx context.Context, tag string, linkID int64) error
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
