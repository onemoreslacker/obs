package storage

import (
	"context"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
)

func (l *LinksSQLService) AddLink(chatID int64, link entities.Link) error {
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

	sql := `INSERT INTO links (id, url, tags, filters) VALUES ($1, $2, $3, $4)`

	result, err := l.pool.Exec(context.Background(), sql, *link.Id, *link.Url, *link.Tags, *link.Filters)
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

func (l *LinksSQLService) GetLinks(chatID int64) (links []entities.Link, err error) {
	if err := l.chatExists(chatID); err != nil {
		return nil, err
	}

	sql := `SELECT id, url, tags, filters FROM tracking_links JOIN links
ON tracking_links.linkID = links.id WHERE chatID = $1`

	rows, err := l.pool.Query(context.Background(), sql, chatID)
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

func (l *LinksSQLService) DeleteLink(chatID int64, url string) error {
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

	sql := `DELETE FROM links WHERE id = $1`

	result, err := l.pool.Exec(context.Background(), sql, linkID)
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

func (l *LinksSQLService) urlExists(url string) error {
	sql := `SELECT id FROM links WHERE url = $1`

	result, err := l.pool.Exec(context.Background(), sql, url)
	if err != nil {
		return err
	}

	if result.RowsAffected() != 0 {
		return scrapperapi.ErrLinkAlreadyExists
	}

	return nil
}

func (l *LinksSQLService) getLinkID(url string) (int64, error) {
	sql := `SELECT id FROM links WHERE url = $1`

	row := l.pool.QueryRow(context.Background(), sql, url)

	var linkID int64

	if err := row.Scan(&linkID); err != nil {
		return 0, err
	}

	return linkID, nil
}
