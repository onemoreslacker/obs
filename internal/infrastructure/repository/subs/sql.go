package subs

import (
	"context"
	"errors"
	"fmt"
	"time"

	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/txs"
	"github.com/jackc/pgx/v5"
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

func (r *SQLRepository) Add(ctx context.Context, chatID, linkID int64) error {
	const query = "INSERT INTO subs (chat_id, link_id) VALUES ($1, $2)"

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, chatID, linkID)
	if err != nil {
		return fmt.Errorf("repo: failed to insert subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrSubscriptionAlreadyExists)
	}

	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, chatID, linkID int64) error {
	const query = "DELETE FROM subs WHERE chat_id = $1 AND link_id = $2"
	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, chatID, linkID)
	if err != nil {
		return fmt.Errorf("repo: failed to delete subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrSubscriptionsNotExists)
	}

	return nil
}

func (r *SQLRepository) GetLinkID(ctx context.Context, url string, chatID int64) (int64, error) {
	const query = "SELECT l.id FROM subs s JOIN links l ON l.id = s.link_id WHERE s.chat_id = $1 AND l.url = $2"

	var linkID int64

	if err := r.db.QueryRow(ctx, query, chatID, url).Scan(&linkID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, sapi.ErrLinkNotExists
		}
		return 0, fmt.Errorf("repo: failed to get link ID: %w", err)
	}

	return linkID, nil
}

func (r *SQLRepository) GetLinksWithChat(ctx context.Context, chatID int64) ([]sapi.LinkResponse, error) {
	const query = `SELECT l.id, l.url, 
COALESCE((SELECT ARRAY_AGG(t.tag) FROM tags t WHERE t.link_id = l.id), '{}') AS tags,
COALESCE((SELECT ARRAY_AGG(f.filter_value) FROM filters f WHERE f.link_id = l.id), '{}') AS filters
FROM subs s JOIN links l ON l.id = s.link_id WHERE s.chat_id = $1 ORDER BY l.updated_at`

	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(ctx, query, chatID)
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

func (r *SQLRepository) GetLinksWithChatActive(ctx context.Context, chatID int64) ([]sapi.LinkResponse, error) {
	const query = `SELECT l.id, l.url, 
COALESCE((SELECT ARRAY_AGG(t.tag) FROM tags t WHERE t.link_id = l.id), '{}') AS tags,
COALESCE((SELECT ARRAY_AGG(f.filter_value) FROM filters f WHERE f.link_id = l.id), '{}') AS filters
FROM subs s JOIN links l ON l.id = s.link_id WHERE s.chat_id = $1 AND l.is_active = TRUE ORDER BY l.updated_at`

	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(ctx, query, chatID)
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
