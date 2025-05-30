package links

import (
	"context"
	"fmt"
	"time"

	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/txs"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *SQLRepository) Add(ctx context.Context, url string) (int64, error) {
	const query = "INSERT INTO links (url, created_at, updated_at) VALUES ($1, $2, $3) RETURNING (id)"

	querier := txs.GetQuerier(ctx, r.db)

	var linkID int64
	if err := querier.QueryRow(ctx, query, url, r.now(), r.now()).Scan(&linkID); err != nil {
		return 0, fmt.Errorf("repo: failed to insert link: %w", err)
	}

	return linkID, nil
}

func (r *SQLRepository) Delete(ctx context.Context, linkID int64) error {
	const query = "DELETE FROM links WHERE id = $1"

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, linkID)
	if err != nil {
		return fmt.Errorf("repo: failed to delete link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrLinkNotExists)
	}

	return nil
}

func (r *SQLRepository) Touch(ctx context.Context, linkID int64) error {
	const query = "UPDATE links SET updated_at = $1 WHERE id = $2"

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, r.now(), linkID)
	if err != nil {
		return fmt.Errorf("repo: failed to update link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return sapi.ErrLinkNotExists
	}

	return nil
}

func (r *SQLRepository) UpdateActivity(ctx context.Context, linkID int64, status bool) error {
	const query = "UPDATE links SET is_active = $1 WHERE id = $2"

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, status, linkID)
	if err != nil {
		return fmt.Errorf("repo: failed to update link activity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return sapi.ErrLinkNotExists
	}

	return nil
}

func (r *SQLRepository) GetBatch(ctx context.Context, batch uint64) ([]sapi.LinkResponse, error) {
	const query = `SELECT l.id, l.url, 
COALESCE((SELECT ARRAY_AGG(t.tag) FROM tags t WHERE t.link_id = l.id), '{}') AS tags,
COALESCE((SELECT ARRAY_AGG(f.filter_value) FROM filters f WHERE f.link_id = l.id), '{}') AS filters
FROM links l ORDER BY l.updated_at LIMIT $1`

	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(ctx, query, batch)
	if err != nil {
		return nil, fmt.Errorf("repo: failed to select links: %w", err)
	}
	defer rows.Close()

	links := make([]sapi.LinkResponse, 0)

	for rows.Next() {
		var (
			link    sapi.LinkResponse
			tags    pgtype.Array[string]
			filters pgtype.Array[string]
		)

		if err := rows.Scan(&link.Id, &link.Url, &tags, &filters); err != nil {
			return nil, fmt.Errorf("repo: failed to scan row: %w", err)
		}

		link.Tags = tags.Elements
		link.Filters = filters.Elements

		links = append(links, link)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("repo: failed to scan rows: %w", err)
	}

	return links, nil
}
