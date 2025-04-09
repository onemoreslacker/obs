package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
)

type LinksRepository interface {
	AddChat(id int64) error
	DeleteChat(id int64) error
	AddLink(id int64, link entities.Link) error
	GetLinks(id int64) ([]entities.Link, error)
	DeleteLink(id int64, url string) error
	GetChatIDs() ([]int64, error)
}

func New(cfg *config.Config, pool *pgxpool.Pool) (LinksRepository, error) {
	slog.Info(
		"LinksService: initialization",
		slog.String("access-type", cfg.Database.AccessType),
	)

	switch cfg.Database.AccessType {
	case "in-memory":
		return NewLinksInMemoryService(), nil
	case "orm":
		return NewLinksOrmService(pool), nil
	case "sql":
		return NewLinksSQLService(pool), nil
	}

	return nil, ErrUnknownDBAccessType
}

func NewPool(cfg *config.Config) (*pgxpool.Pool, error) {
	conn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?target_session_attrs=read-write&sslmode=disable",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	connConfig, err := pgxpool.ParseConfig(conn)
	if err != nil {
		return nil, err
	}

	connConfig.MaxConns = 32
	connConfig.MaxConnIdleTime = time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), connConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pool, nil
}
