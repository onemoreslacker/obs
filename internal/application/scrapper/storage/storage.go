package storage

import (
	"context"
	"errors"
	"fmt"

	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
)

type ChatsRepository interface {
	Add(ctx context.Context, chatID int64) error
	Delete(ctx context.Context, chatID int64) error
	ExistsID(ctx context.Context, chatID int64) error
	GetIDs(ctx context.Context) ([]int64, error)
}

type LinksRepository interface {
	Add(ctx context.Context, url string) (int64, error)
	Delete(ctx context.Context, linkID int64) error
	Touch(ctx context.Context, linkID int64) error
	UpdateActivity(ctx context.Context, linkID int64, status bool) error
	GetBatch(ctx context.Context, batch uint64) ([]sapi.LinkResponse, error)
}

type SubsRepository interface {
	Add(ctx context.Context, chatID, linkID int64) error
	Delete(ctx context.Context, chatID, linkID int64) error
	GetLinkID(ctx context.Context, url string, chatID int64) (int64, error)
	GetLinksWithChat(ctx context.Context, chatID int64) ([]sapi.LinkResponse, error)
	GetLinksWithChatActive(ctx context.Context, chatID int64) ([]sapi.LinkResponse, error)
}

type TagsRepository interface {
	Add(ctx context.Context, tag string, linkID int64) error
}

type FiltersRepository interface {
	Add(ctx context.Context, filter string, linkID int64) error
}

type Transactor interface {
	WithTransaction(ctx context.Context, txFunc func(ctx context.Context) error) (err error)
}

type Storage struct {
	chats   ChatsRepository
	links   LinksRepository
	subs    SubsRepository
	tags    TagsRepository
	filters FiltersRepository
	tx      Transactor
}

func New(
	chats ChatsRepository,
	links LinksRepository,
	subs SubsRepository,
	tags TagsRepository,
	filters FiltersRepository,
	tx Transactor,
) *Storage {
	return &Storage{
		chats:   chats,
		links:   links,
		subs:    subs,
		tags:    tags,
		filters: filters,
		tx:      tx,
	}
}

func (s *Storage) AddChat(ctx context.Context, chatID int64) error {
	return s.chats.Add(ctx, chatID)
}

func (s *Storage) DeleteChat(ctx context.Context, chatID int64) error {
	return s.chats.Delete(ctx, chatID)
}

func (s *Storage) ExistsChat(ctx context.Context, chatID int64) error {
	return s.chats.ExistsID(ctx, chatID)
}

func (s *Storage) AddLink(ctx context.Context, link sapi.AddLinkRequest, chatID int64) (int64, error) {
	var linkID int64

	err := s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		_, err := s.subs.GetLinkID(ctx, link.Link, chatID)
		if !errors.Is(err, sapi.ErrLinkNotExists) {
			return fmt.Errorf("storage: %w", sapi.ErrLinkAlreadyExists)
		}

		id, err := s.links.Add(ctx, link.Link)
		if err != nil {
			return fmt.Errorf("storage: failed to create link: %w", err)
		}
		linkID = id

		if err = s.subs.Add(ctx, chatID, linkID); err != nil {
			return fmt.Errorf("storage: failed to create subscription: %w", err)
		}

		for _, tag := range link.Tags {
			if err = s.tags.Add(ctx, tag, linkID); err != nil {
				return fmt.Errorf("storage: failed to add tag: %w", err)
			}
		}

		for _, filter := range link.Filters {
			if err = s.filters.Add(ctx, filter, linkID); err != nil {
				return fmt.Errorf("storage: failed to add filter: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("storage: transaction failed: %w", err)
	}

	return linkID, nil
}

func (s *Storage) DeleteLink(ctx context.Context, link sapi.RemoveLinkRequest, chatID int64) error {
	linkID, err := s.subs.GetLinkID(ctx, link.Link, chatID)
	if err != nil {
		return fmt.Errorf("storage: %w", err)
	}

	if err = s.links.Delete(ctx, linkID); err != nil {
		return fmt.Errorf("storage: failed to delete link: %w", err)
	}

	return nil
}

func (s *Storage) GetLinks(ctx context.Context, batch uint64) ([]sapi.LinkResponse, error) {
	return s.links.GetBatch(ctx, batch)
}

func (s *Storage) TouchLink(ctx context.Context, linkID int64) error {
	return s.links.Touch(ctx, linkID)
}

func (s *Storage) GetLinksWithChat(ctx context.Context, chatID int64) ([]sapi.LinkResponse, error) {

	return s.subs.GetLinksWithChat(ctx, chatID)
}

func (s *Storage) GetLinksWithChatActive(ctx context.Context, chatID int64) ([]sapi.LinkResponse, error) {
	return s.subs.GetLinksWithChatActive(ctx, chatID)
}

func (s *Storage) UpdateLinkActivity(ctx context.Context, linkID int64, status bool) error {
	return s.links.UpdateActivity(ctx, linkID, status)
}

func (s *Storage) GetChatIDs(ctx context.Context) ([]int64, error) {
	return s.chats.GetIDs(ctx)
}

func (s *Storage) GetLinkID(ctx context.Context, url string, chatID int64) (int64, error) {
	return s.subs.GetLinkID(ctx, url, chatID)
}
