package storage

import (
	"context"
	"fmt"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/updater"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinksSQLService struct {
	pool *pgxpool.Pool
}

func NewLinksSQLService(pool *pgxpool.Pool) *LinksSQLService {
	return &LinksSQLService{
		pool: pool,
	}
}

// AddChat registers a new chat in the system.
func (l *LinksSQLService) AddChat(chatID int64) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	sql := `INSERT INTO chats (id) VALUES ($1)`

	result, err := l.pool.Exec(context.Background(), sql, chatID)
	if err != nil {
		return fmt.Errorf("failed to insert chat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return scrapperapi.ErrChatAlreadyExists
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteChat removes a chat and its associated tracking links via cascade.
func (l *LinksSQLService) DeleteChat(chatID int64) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	sql := `DELETE FROM chats WHERE id = $1`

	result, err := tx.Exec(context.Background(), sql, chatID)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return scrapperapi.ErrChatNotFound
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// AddLink adds a new tracking link to a chat.
func (l *LinksSQLService) AddLink(chatID int64, link models.Link) error {
	if link.Id == nil || link.Url == nil || link.Tags == nil || link.Filters == nil {
		return scrapperapi.ErrAddLinkInvalidLink
	}

	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	if err := l.chatExists(tx, chatID); err != nil {
		return err
	}

	sql := `INSERT INTO links (id, url, tags, filters) VALUES ($1, $2, $3, $4)`

	result, err := tx.Exec(context.Background(), sql, *link.Id, *link.Url, *link.Tags, *link.Filters)
	if err != nil {
		return fmt.Errorf("failed to insert link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return scrapperapi.ErrAddLinkFailed
	}

	if err := l.addTrackingLink(tx, chatID, *link.Id); err != nil {
		return err
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetChatLinks retrieves all active links for a specific chat.
func (l *LinksSQLService) GetChatLinks(chatID int64, includeAll bool) ([]models.Link, error) {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := l.chatExists(tx, chatID); err != nil {
		return nil, err
	}

	var sql string

	if includeAll {
		sql = `SELECT id, url, tags, filters FROM tracking_links JOIN links
ON tracking_links.link_id = links.id WHERE chat_id = $1`
	} else {
		sql = `SELECT id, url, tags, filters FROM tracking_links JOIN links
ON tracking_links.link_id = links.id WHERE chat_id = $1 AND is_activity_recorded`
	}

	rows, err := l.pool.Query(context.Background(), sql, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to select chat links: %w", err)
	}
	defer rows.Close()

	links := make([]models.Link, 0)

	for rows.Next() {
		var (
			id      int64
			url     string
			tags    []string
			filters []string
		)

		if err := rows.Scan(&id, &url, &tags, &filters); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		links = append(links, models.NewLink(id, url, tags, filters))
	}

	if err := tx.Commit(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return links, nil
}

func (l *LinksSQLService) DeleteLink(chatID int64, url string) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	linkID, err := l.getLinkID(tx, url)
	if err != nil {
		return err
	}

	sql := `DELETE FROM tracking_links WHERE chat_id = $1 AND link_id = $2`

	result, err := tx.Exec(context.Background(), sql, chatID, linkID)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return scrapperapi.ErrLinkNotFound
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TouchLink updates the timestamp of a link.
func (l *LinksSQLService) TouchLink(linkID int64) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	sql := `UPDATE links SET updated_at = NOW() WHERE id = $1`

	result, err := tx.Exec(context.Background(), sql, linkID)
	if err != nil {
		return fmt.Errorf("failed to update link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return updater.ErrLinkUpdateFailed
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateLinkActivity updates the activity tracking status of a link.
func (l *LinksSQLService) UpdateLinkActivity(linkID int64, status bool) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	sql := `UPDATE links SET is_activity_recorded = $1 WHERE id = $2`

	result, err := l.pool.Exec(context.Background(), sql, status, linkID)
	if err != nil {
		return fmt.Errorf("failed to update link activity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return scrapperapi.ErrLinkNotFound
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetLinks returns all links being tracked.
func (l *LinksSQLService) GetLinks(batchSize uint64) ([]models.Link, error) {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	sql := `SELECT id, url, tags, filters FROM links ORDER BY updated_at LIMIT $1 `

	rows, err := l.pool.Query(context.Background(), sql, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to select links: %w", err)
	}
	defer rows.Close()

	links := make([]models.Link, 0, batchSize)

	for rows.Next() {
		var (
			id      int64
			url     string
			tags    []string
			filters []string
		)

		if err := rows.Scan(&id, &url, &tags, &filters); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		links = append(links, models.NewLink(id, url, tags, filters))
	}

	if err := tx.Commit(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return links, nil
}

// GetChatsIDs retrieves all registered chat IDs.
func (l *LinksSQLService) GetChatsIDs() (ids []int64, err error) {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	sql := `SELECT id FROM chats`

	rows, err := l.pool.Query(context.Background(), sql)
	if err != nil {
		return nil, fmt.Errorf("failed to select chats ids: %w", err)
	}

	ids = make([]int64, 0)

	for rows.Next() {
		var id int64

		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		ids = append(ids, id)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ids, nil
}

// chatExists checks for chat existence using proper parameter binding.
func (l *LinksSQLService) chatExists(tx pgx.Tx, chatID int64) error {
	sql := `SELECT EXISTS (SELECT 1 FROM chats WHERE id = $1)`

	var exists bool
	if err := tx.QueryRow(context.Background(), sql, chatID).Scan(&exists); err != nil {
		return fmt.Errorf("failed to execute existence check: %w", err)
	}

	if !exists {
		return scrapperapi.ErrChatNotFound
	}

	return nil
}

// getLinkID retrieves link's id.
func (l *LinksSQLService) getLinkID(tx pgx.Tx, url string) (int64, error) {
	sql := `SELECT id FROM links WHERE url = $1`

	var linkID int64

	if err := tx.QueryRow(context.Background(), sql, url).Scan(&linkID); err != nil {
		return 0, fmt.Errorf("failed to get link id: %w", err)
	}

	return linkID, nil
}

// addTrackingLink adds link and associated chat to the tracking_links table.
func (l *LinksSQLService) addTrackingLink(tx pgx.Tx, chatID, linkID int64) error {
	sql := `INSERT INTO tracking_links VALUES ($1, $2)`

	_, err := tx.Exec(context.Background(), sql, chatID, linkID)
	if err != nil {
		return fmt.Errorf("failed to create tracking relationship: %w", err)
	}

	return nil
}
