package db

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func New(cfg config.Database) (*pgxpool.Pool, error) {
	pool, err := newPool(cfg)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	connConfig := pool.Config().ConnConfig

	if err := ApplyMigrations(connConfig); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: %w", err)
	}

	return pool, nil
}

func newPool(cfg config.Database) (*pgxpool.Pool, error) {
	connConfig, err := pgxpool.ParseConfig(cfg.ToDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
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

//go:embed migrations/*.sql
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
