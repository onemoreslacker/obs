package links

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/txs"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *SquirrelRepository) Add(ctx context.Context, url string) (int64, error) {
	query, args, err := r.sb.Insert("links").
		Columns("url").
		Values(url).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("repo: failed to build insert query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	var linkID int64
	if err := querier.QueryRow(ctx, query, args...).Scan(&linkID); err != nil {
		return 0, fmt.Errorf("repo: failed to insert link: %w", err)
	}

	return linkID, nil
}

func (r *SquirrelRepository) Delete(ctx context.Context, linkID int64) error {
	query, args, err := r.sb.Delete("links").
		Where(sq.Eq{
			"id": linkID,
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("repo: failed to build delte query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("repo: failed to delete link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrLinkNotExists)
	}

	return nil
}

func (r *SquirrelRepository) Touch(ctx context.Context, linkID int64) error {
	query, args, err := r.sb.Update("links").
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": linkID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("repo: failed to build update query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("repo: failed to update link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return sapi.ErrLinkNotExists
	}

	return nil
}

func (r *SquirrelRepository) UpdateActivity(ctx context.Context, linkID int64, status bool) error {
	query, args, err := r.sb.Update("links").
		Set("is_active", status).
		Where(sq.Eq{"id": linkID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("repo: failed to build update query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("repo: failed to update link activity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return sapi.ErrLinkNotExists
	}

	return nil
}

func (r *SquirrelRepository) GetBatch(ctx context.Context, batch uint64) ([]sapi.LinkResponse, error) {
	query, args, err := r.sb.Select(
		"l.id",
		"l.url",
		"COALESCE((SELECT ARRAY_AGG(t.tag) FROM tags t WHERE t.link_id = l.id), '{}') AS tags",
		"COALESCE((SELECT ARRAY_AGG(f.filter_value) FROM filters f WHERE f.link_id = l.id), '{}') AS filters",
	).
		From("links l").
		OrderBy("l.updated_at").
		Limit(batch).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("repo: failed to build select query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(ctx, query, args...)
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
