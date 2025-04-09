package storage

import (
	"context"
	"log/slog"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
)

func (l *LinksSQLService) AddChat(chatID int64) error {
	sql := `INSERT INTO chats (id) VALUES ($1)`

	result, err := l.pool.Exec(context.Background(), sql, chatID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return scrapperapi.ErrChatAlreadyExists
	}

	slog.Info(
		"LinksOrmService: chat is registered",
		slog.Int64("chatID", chatID),
	)

	return nil
}

func (l *LinksSQLService) DeleteChat(chatID int64) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(context.Background())
		}
	}()

	sql := `DELETE FROM chats WHERE id = $1`

	result, err := l.pool.Exec(context.Background(), sql, chatID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return scrapperapi.ErrChatNotFound
	}

	return tx.Commit(context.Background())
}

func (l *LinksSQLService) GetChatIDs() (ids []int64, err error) {
	sql := `SELECT id FROM chats`

	rows, err := l.pool.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}

	ids = make([]int64, 0)

	for rows.Next() {
		var id int64

		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (l *LinksSQLService) chatExists(chatID int64) error {
	sql := `SELECT id FROM chats WHERE id = $1`

	exists, err := l.pool.Exec(context.Background(), sql, chatID)
	if err != nil {
		return err
	}

	if exists.RowsAffected() == 0 {
		return scrapperapi.ErrChatNotFound
	}

	return nil
}
