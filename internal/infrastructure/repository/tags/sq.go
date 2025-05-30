package tags

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/txs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SquirrelRepository struct {
	db  *pgxpool.Pool
	sb  sq.StatementBuilderType
	now func() time.Time
}

func NewSquirrelRepository(pool *pgxpool.Pool) *SquirrelRepository {
	return &SquirrelRepository{
		db:  pool,
		sb:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		now: time.Now,
	}
}

func (r *SquirrelRepository) Add(ctx context.Context, tag string, linkID int64) error {
	sql, args, err := r.sb.Insert("tags").
		Columns("tag", "link_id").
		Values(tag, linkID).
		ToSql()
	if err != nil {
		return fmt.Errorf("repo: failed to build insert query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("repo: failed to insert tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrTagAlreadyExists)
	}

	return nil
}
