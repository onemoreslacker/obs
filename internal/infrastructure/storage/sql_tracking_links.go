package storage

import (
	"context"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
)

func (l *LinksSQLService) addTrackingLink(chatID, linkID int64) error {
	sql := `INSERT INTO tracking_links VALUES ($1, $2)`

	result, err := l.pool.Exec(context.Background(), sql, chatID, linkID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return scrapperapi.ErrAddLinkFailed
	}

	return nil
}
