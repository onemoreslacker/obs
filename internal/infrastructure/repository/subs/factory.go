package subs

import (
	"context"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Add(ctx context.Context, chatID, linkID int64) error
	Delete(ctx context.Context, chatID, linkID int64) error
	GetLinkID(ctx context.Context, url string, chatID int64) (int64, error)
	GetLinksWithChat(ctx context.Context, chatID int64) ([]sapi.LinkResponse, error)
	GetLinksWithChatActive(ctx context.Context, chatID int64) ([]sapi.LinkResponse, error)
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
