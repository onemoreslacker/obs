package storage

import (
	"context"
	"log/slog"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
)

func (l *LinksOrmService) AddChat(chatID int64) error {
	sql, args, err := l.sb.Insert("chats").
		Columns("id").
		Values(chatID).
		ToSql()
	if err != nil {
		return err
	}

	result, err := l.pool.Exec(context.Background(), sql, args...)
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

func (l *LinksOrmService) DeleteChat(chatID int64) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(context.Background())
		}
	}()

	sql, args, err := l.sb.Delete("chats").
		Where("id = ?", chatID).
		ToSql()
	if err != nil {
		return err
	}

	result, err := l.pool.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return scrapperapi.ErrChatNotFound
	}

	return tx.Commit(context.Background())
}

func (l *LinksOrmService) GetChatIDs() (ids []int64, err error) {
	sql, args, err := l.sb.Select("id").
		From("chats").
		ToSql()
	if err != nil {
		return
	}

	rows, err := l.pool.Query(context.Background(), sql, args...)
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

func (l *LinksOrmService) chatExists(chatID int64) error {
	sql, args, err := l.sb.Select("id").
		From("chats").
		Where("id = ?", chatID).
		ToSql()
	if err != nil {
		return err
	}

	exists, err := l.pool.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	if exists.RowsAffected() == 0 {
		return scrapperapi.ErrChatNotFound
	}

	return nil
}
