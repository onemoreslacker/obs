package chats

import (
	"context"
	"database/sql"
	"errors"
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

func NewSquirrelRepository(pool *pgxpool.Pool) Repository {
	return &SquirrelRepository{
		db:  pool,
		sb:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		now: time.Now,
	}
}

func (r *SquirrelRepository) Add(ctx context.Context, chatID int64) error {
	query, args, err := r.sb.Insert("chats").
		Columns("id", "created_at", "updated_at").
		Values(chatID, r.now(), r.now()).
		ToSql()
	if err != nil {
		return fmt.Errorf("repo: failed to build insert query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("repo: failed to insert chat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrChatAlreadyExists)
	}

	return nil
}

func (r *SquirrelRepository) Delete(ctx context.Context, chatID int64) error {
	query, args, err := r.sb.Delete("chats").
		Where(sq.Eq{"id": chatID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("repo: failed to build delete query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("repo: failed to delete chat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrChatNotExists)
	}

	return nil
}

func (r *SquirrelRepository) ExistsID(ctx context.Context, chatID int64) error {
	query, args, err := r.sb.Select("1").
		From("chats").
		Where(sq.Eq{"id": chatID}).ToSql()
	if err != nil {
		return fmt.Errorf("repo: failed to build select query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	if err := querier.QueryRow(ctx, query, args...).Scan(&chatID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("repo: chat does not exist: %w", sapi.ErrChatNotExists)
		}
		return fmt.Errorf("repo: failed to select chat id: %w", err)
	}

	return nil
}

func (r *SquirrelRepository) GetIDs(ctx context.Context) ([]int64, error) {
	query, args, err := r.sb.Select("id").
		From("chats").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("repo: failed to build select query: %w", err)
	}

	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("repo: failed to select chats ids: %w", err)
	}
	defer rows.Close()

	chatIDs := make([]int64, 0)

	for rows.Next() {
		var chatID int64

		if err := rows.Scan(&chatID); err != nil {
			return nil, fmt.Errorf("repo: failed to scan row: %w", err)
		}

		chatIDs = append(chatIDs, chatID)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("repo: failed to scan rows: %w", err)
	}

	return chatIDs, nil
}
