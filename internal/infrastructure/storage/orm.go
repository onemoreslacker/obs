package storage

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LinksOrmService implements LinksService using PostgreSQL with Squirrel SQL builder.
type LinksOrmService struct {
	pool *pgxpool.Pool
	sb   sq.StatementBuilderType
}

// NewLinksOrmService creates a new ORM-based links service.
func NewLinksOrmService(pool *pgxpool.Pool) LinksService {
	return &LinksOrmService{
		pool: pool,
		sb:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// AddChat registers a new chat in the system.
func (l *LinksOrmService) AddChat(chatID int64) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	sql, args, err := l.sb.Insert("chats").
		Columns("id").
		Values(chatID).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	result, err := l.pool.Exec(context.Background(), sql, args...)
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
func (l *LinksOrmService) DeleteChat(chatID int64) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	sql, args, err := l.sb.Delete("chats").
		Where(sq.Eq{"id": chatID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := tx.Exec(context.Background(), sql, args...)
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
func (l *LinksOrmService) AddLink(chatID int64, link models.Link) error {
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

	sql, args, err := l.sb.Insert("links").
		Columns("id", "url", "tags", "filters").
		Values(*link.Id, *link.Url, *link.Tags, *link.Filters).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	result, err := tx.Exec(context.Background(), sql, args...)
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
func (l *LinksOrmService) GetChatLinks(chatID int64, includeAll bool) ([]models.Link, error) {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	if err := l.chatExists(tx, chatID); err != nil {
		return nil, err
	}

	query := l.sb.Select("links.id", "links.url", "links.tags", "links.filters").
		From("tracking_links").
		Join("links ON tracking_links.link_id = links.id").
		Where(sq.Eq{
			"tracking_links.chat_id": chatID,
		})

	if !includeAll {
		query = query.Where(sq.Eq{"is_activity_recorded": true})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := l.pool.Query(context.Background(), sql, args...)
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

// DeleteLink removes a tracking relationship between a chat and a link.
func (l *LinksOrmService) DeleteLink(chatID int64, url string) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	linkID, err := l.getLinkID(tx, url)
	if err != nil {
		return err
	}

	sql, args, err := l.sb.Delete("tracking_links").
		Where(sq.Eq{
			"chat_id": chatID,
			"link_id": linkID,
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := tx.Exec(context.Background(), sql, args...)
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
func (l *LinksOrmService) TouchLink(linkID int64) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	sql, args, err := l.sb.Update("links").
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": linkID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := tx.Exec(context.Background(), sql, args...)
	if err != nil {
		return fmt.Errorf("failed to update link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return scrapperapi.ErrLinkNotFound
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateLinkActivity updates the activity tracking status of a link.
func (l *LinksOrmService) UpdateLinkActivity(linkID int64, status bool) error {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	sql, args, err := l.sb.Update("links").
		Set("is_activity_recorded", status).
		Where(sq.Eq{"id": linkID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := l.pool.Exec(context.Background(), sql, args...)
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
func (l *LinksOrmService) GetLinks(batchSize uint64) ([]models.Link, error) {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	sql, args, err := l.sb.Select("id, url, tags, filters").
		From("links").
		OrderBy("updated_at").
		Limit(batchSize).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := l.pool.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to select links: %w", err)
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

// GetChatsIDs retrieves all registered chat IDs.
func (l *LinksOrmService) GetChatsIDs() ([]int64, error) {
	tx, err := l.pool.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	sql, args, err := l.sb.Select("id").
		From("chats").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := l.pool.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to select chats ids: %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0)

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
func (l *LinksOrmService) chatExists(tx pgx.Tx, chatID int64) error {
	sql, args, err := l.sb.Select("1").
		From("chats").
		Where(sq.Eq{"id": chatID}).
		Prefix("SELECT EXISTS (").
		Suffix(")").ToSql()
	if err != nil {
		return fmt.Errorf("failed to build existence query: %w", err)
	}

	var exists bool
	if err := tx.QueryRow(context.Background(), sql, args...).Scan(&exists); err != nil {
		return fmt.Errorf("failed to execute existence check: %w", err)
	}

	if !exists {
		return scrapperapi.ErrChatNotFound
	}

	return nil
}

// getLinkID retrieves link's id.
func (l *LinksOrmService) getLinkID(tx pgx.Tx, url string) (int64, error) {
	sql, args, err := l.sb.Select("id").
		From("links").
		Where(sq.Eq{"url": url}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build select query: %w", err)
	}

	var linkID int64

	if err := tx.QueryRow(context.Background(), sql, args...).Scan(&linkID); err != nil {
		return 0, fmt.Errorf("failed to get link ID: %w", err)
	}

	return linkID, nil
}

// addTrackingLink adds link and associated chat to the tracking_links table.
func (l *LinksOrmService) addTrackingLink(tx pgx.Tx, chatID, linkID int64) error {
	sql, args, err := l.sb.Insert("tracking_links").
		Columns("chat_id", "link_id").
		Values(chatID, linkID).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = tx.Exec(context.Background(), sql, args...)
	if err != nil {
		return fmt.Errorf("failed to create tracking relationship: %w", err)
	}

	return nil
}
