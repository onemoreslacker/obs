package chats

import (
	"context"
	"database/sql"
	"errors"
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

func (r *SQLRepository) Add(ctx context.Context, chatID int64) error {
	const query = "INSERT INTO chats (id, created_at, updated_at) VALUES ($1, $2, $3)"

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, chatID, r.now(), r.now())
	if err != nil {
		return fmt.Errorf("repo: failed to insert chat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrChatAlreadyExists)
	}

	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, chatID int64) error {
	const query = "DELETE FROM chats WHERE id = $1"

	querier := txs.GetQuerier(ctx, r.db)

	result, err := querier.Exec(ctx, query, chatID)
	if err != nil {
		return fmt.Errorf("repo: failed to delete chat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repo: %w", sapi.ErrChatNotExists)
	}

	return nil
}

func (r *SQLRepository) ExistsID(ctx context.Context, chatID int64) error {
	const query = "SELECT 1 FROM chats WHERE id = $1"

	querier := txs.GetQuerier(ctx, r.db)

	if err := querier.QueryRow(ctx, query, chatID).Scan(&chatID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("repo: chat does not exist: %w", sapi.ErrChatNotExists)
		}
		return fmt.Errorf("repo: failed to select chat id: %w", err)
	}

	return nil
}

func (r *SQLRepository) GetIDs(ctx context.Context) ([]int64, error) {
	const query = "SELECT id FROM chats"

	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(ctx, query)
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
