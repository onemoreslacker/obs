package tags

import (
	"context"
	"fmt"
	"time"

	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/txs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	db  *pgxpool.Pool
	now func() time.Time
}

func NewSQLRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{
		db:  pool,
		now: time.Now,
	}
}

func (r *SQLRepository) Add(ctx context.Context, tag string, linkID int64) error {
	const query = "INSERT INTO tags (tag, link_id) VALUES ($1, $2)"

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, tag, linkID)
	if err != nil {
		return fmt.Errorf("repo: failed to insert tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrTagAlreadyExists)
	}

	return nil
}
