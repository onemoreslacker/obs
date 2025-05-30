package subs

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/txs"
	"github.com/jackc/pgx/v5"
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

func (r *SquirrelRepository) Add(ctx context.Context, chatID, linkID int64) error {
	sql, args, err := r.sb.Insert("subs").
		Columns("chat_id", "link_id").
		Values(chatID, linkID).
		ToSql()
	if err != nil {
		return fmt.Errorf("repo: failed to build insert query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("repo: failed to insert subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrSubscriptionAlreadyExists)
	}

	return nil
}

func (r *SquirrelRepository) Delete(ctx context.Context, chatID, linkID int64) error {
	sql, args, err := r.sb.Delete("subs").
		Where(sq.Eq{
			"chat_id": chatID,
			"link_id": linkID,
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("repo: failed to build delete query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("repo: failed to delete subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrSubscriptionsNotExists)
	}

	return nil
}

func (r *SquirrelRepository) GetLinkID(ctx context.Context, url string, chatID int64) (int64, error) {
	query, args, err := r.sb.Select("l.id").
		From("subs s").
		Join("links l ON l.id = s.link_id").
		Where(sq.Eq{
			"s.chat_id": chatID,
			"l.url":     url,
		}).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("repo: failed to build query: %w", err)
	}

	var linkID int64

	if err = r.db.QueryRow(ctx, query, args...).Scan(&linkID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, sapi.ErrLinkNotExists
		}
		return 0, fmt.Errorf("repo: failed to get link ID: %w", err)
	}

	return linkID, nil
}

func (r *SquirrelRepository) GetLinksWithChat(ctx context.Context, chatID int64) ([]sapi.LinkResponse, error) {
	sql, args, err := r.sb.Select(
		"l.id",
		"l.url",
		"COALESCE((SELECT ARRAY_AGG(t.tag) FROM tags t WHERE t.link_id = l.id), '{}') AS tags",
		"COALESCE((SELECT ARRAY_AGG(f.filter_value) FROM filters f WHERE f.link_id = l.id), '{}') AS filters",
	).
		From("subs s").
		Join("links l ON s.link_id = l.id").
		Where(sq.Eq{
			"s.chat_id": chatID,
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("repo: failed to build select query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(ctx, sql, args...)
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

func (r *SquirrelRepository) GetLinksWithChatActive(ctx context.Context, chatID int64) ([]sapi.LinkResponse, error) {
	sql, args, err := r.sb.Select(
		"l.id",
		"l.url",
		"COALESCE((SELECT ARRAY_AGG(t.tag) FROM tags t WHERE t.link_id = l.id), '{}') AS tags",
		"COALESCE((SELECT ARRAY_AGG(f.filter_value) FROM filters f WHERE f.link_id = l.id), '{}') AS filters",
	).
		From("subs s").
		Join("links l ON s.link_id = l.id").
		Where(sq.Eq{
			"s.chat_id":   chatID,
			"l.is_active": true,
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("repo: failed to build select query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(ctx, sql, args...)
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
