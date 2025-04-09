package storage

import (
	"context"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
)

func (l *LinksOrmService) AddLink(chatID int64, link entities.Link) error {
	if link.Id == nil || link.Url == nil || link.Tags == nil || link.Filters == nil {
		return scrapperapi.ErrAddLinkInvalidLink
	}

	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(context.Background())
		}
	}()

	if err := l.chatExists(chatID); err != nil {
		return err
	}

	if err := l.urlExists(*link.Url); err != nil {
		return err
	}

	sql, args, err := l.sb.Insert("links").
		Columns("id", "url", "tags", "filters").
		Values(*link.Id, *link.Url, *link.Tags, *link.Filters).
		ToSql()
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

	if err := l.addTrackingLink(chatID, *link.Id); err != nil {
		return err
	}

	return tx.Commit(context.Background())
}

func (l *LinksOrmService) GetLinks(chatID int64) (links []entities.Link, err error) {
	if err := l.chatExists(chatID); err != nil {
		return nil, err
	}

	sql, args, err := l.sb.Select("id, url, tags, filters").
		From("tracking_links").
		Join("links ON tracking_links.linkID = links.id").
		Where("chatID = ?", chatID).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := l.pool.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links = make([]entities.Link, 0)

	for rows.Next() {
		var (
			id      int64
			url     string
			tags    []string
			filters []string
		)

		if err := rows.Scan(&id, &url, &tags, &filters); err != nil {
			return nil, err
		}

		links = append(links, entities.NewLink(id, url, tags, filters))
	}

	return links, nil
}

func (l *LinksOrmService) DeleteLink(chatID int64, url string) error {
	if err := l.chatExists(chatID); err != nil {
		return err
	}

	linkID, err := l.getLinkID(url)
	if err != nil {
		return err
	}

	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(context.Background())
		}
	}()

	sql, args, err := l.sb.Delete("links").
		Where("id = ?", linkID).
		ToSql()
	if err != nil {
		return err
	}

	result, err := l.pool.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() != 1 {
		return scrapperapi.ErrLinkNotFound
	}

	if err := tx.Commit(context.Background()); err != nil {
		return err
	}

	return nil
}

func (l *LinksOrmService) urlExists(url string) error {
	sql, args, err := l.sb.Select("id").
		From("links").
		Where("url = ?", url).
		ToSql()
	if err != nil {
		return err
	}

	result, err := l.pool.Exec(context.Background(), sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() != 0 {
		return scrapperapi.ErrLinkAlreadyExists
	}

	return nil
}

func (l *LinksOrmService) getLinkID(url string) (int64, error) {
	sql, args, err := l.sb.Select("id").
		From("links").
		Where("url = ?", url).
		ToSql()
	if err != nil {
		return 0, err
	}

	row := l.pool.QueryRow(context.Background(), sql, args...)

	var linkID int64

	if err := row.Scan(&linkID); err != nil {
		return 0, err
	}

	return linkID, nil
}
