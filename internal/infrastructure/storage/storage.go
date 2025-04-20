package storage

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"

	"github.com/pressly/goose/v3"
)

type LinksService interface {
	AddChat(chatID int64) error
	DeleteChat(chatID int64) error
	AddLink(chatID int64, link models.Link) error
	GetChatLinks(chatID int64, includeAll bool) ([]models.Link, error)
	DeleteLink(chatID int64, url string) error
	GetLinks(batchSize uint64) ([]models.Link, error)
	TouchLink(linkID int64) error
	UpdateLinkActivity(linkID int64, status bool) error
	GetChatsIDs() ([]int64, error)
}

func New(cfg *config.Config, pool *pgxpool.Pool) (LinksService, error) {
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

func NewPoolWithMigrations(cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := NewPool(cfg)
	if err != nil {
		return nil, err
	}

	connConfig := pool.Config().ConnConfig

	if err := ApplyMigrations(connConfig); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	return pool, nil
}

//go:embed migrations/00[1-3]_*.sql
var embedMigrations embed.FS

func ApplyMigrations(cfg *pgx.ConnConfig) error {
	goose.SetBaseFS(embedMigrations)

	db := stdlib.OpenDB(*cfg)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
