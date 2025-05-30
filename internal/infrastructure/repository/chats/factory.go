package chats

import (
	"context"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Add(ctx context.Context, chatID int64) error
	Delete(ctx context.Context, chatID int64) error
	ExistsID(ctx context.Context, chatID int64) error
	GetIDs(ctx context.Context) ([]int64, error)
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

func New(cfg config.Database, pool *pgxpool.Pool, opts ...Option) Repository {
	var repo Repository

	switch cfg.Access {
	case config.Orm:
		repo = NewSquirrelRepository(pool)
	case config.Sql:
		repo = NewSQLRepository(pool)
	}

	for _, opt := range opts {
		opt(repo)
	}

	return repo
}
