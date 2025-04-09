package storage

import (
	"context"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
)

func (l *LinksOrmService) addTrackingLink(chatID, linkID int64) error {
	sql, args, err := l.sb.Insert("tracking_links").
		Values(chatID, linkID).ToSql()
	if err != nil {
		return err
	}

	result, err := l.pool.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return scrapperapi.ErrAddLinkFailed
	}

	return nil
}
