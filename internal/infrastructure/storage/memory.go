package storage

import (
	"context"
	"log/slog"
	"sync"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

// LinksInMemoryService stores links tracked by user and guarded by mutex.
type LinksInMemoryService struct {
	links map[int64]map[string]models.Link
	mu    sync.Mutex
}

// NewLinksInMemoryService implements a new LinksInMemoryService entity.
func NewLinksInMemoryService() *LinksInMemoryService {
	return &LinksInMemoryService{
		links: make(map[int64]map[string]models.Link),
	}
}

// AddChat creates a map of links for a new chat.
func (r *LinksInMemoryService) AddChat(_ context.Context, chatID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.links[chatID]; exists {
		return scrapperapi.ErrChatAlreadyExists
	}

	r.links[chatID] = make(map[string]models.Link)

	return nil
}

// DeleteChat deletes all links for the provided chat.
func (r *LinksInMemoryService) DeleteChat(_ context.Context, chatID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.links[chatID]; !exists {
		return scrapperapi.ErrChatNotFound
	}

	delete(r.links, chatID)

	return nil
}

// AddLink adds new tracking link.
func (r *LinksInMemoryService) AddLink(_ context.Context, chatID int64, link models.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, exists := r.links[chatID]
	if !exists {
		return scrapperapi.ErrChatNotFound
	}

	if link.Url == nil {
		return scrapperapi.ErrAddLinkInvalidLink
	}

	slog.Info(
		"repository: add link",
		slog.String("link", *link.Url),
	)

	if _, exists := entries[*link.Url]; exists {
		return scrapperapi.ErrLinkAlreadyExists
	}

	entries[*link.Url] = link

	return nil
}

// GetChatLinks retrieves links attached to the chat id.
func (r *LinksInMemoryService) GetChatLinks(_ context.Context, chatID int64, _ bool) ([]models.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, exists := r.links[chatID]
	if !exists {
		return nil, scrapperapi.ErrChatNotFound
	}

	links := make([]models.Link, 0, len(entries))

	for _, link := range entries {
		links = append(links, link)
	}

	return links, nil
}

// DeleteLink deletes link attached to the chat id.
func (r *LinksInMemoryService) DeleteLink(_ context.Context, chatID int64, url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.links[chatID]
	if !exists {
		return scrapperapi.ErrChatNotFound
	}

	entries := r.links[chatID]

	if _, exists := entries[url]; !exists {
		return scrapperapi.ErrLinkNotFound
	}

	delete(entries, url)

	return nil
}

// GetChatIDs returns all the registered chat IDs.
func (r *LinksInMemoryService) GetChatIDs(_ context.Context) ([]int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids := make([]int64, 0, len(r.links))
	for id := range r.links {
		ids = append(ids, id)
	}

	return ids, nil
}

// GetLinks returns all tracking links.
func (r *LinksInMemoryService) GetLinks(_ context.Context, batchSize uint64) ([]models.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	links := make([]models.Link, 0, batchSize)

	for id := range r.links {
		for _, link := range r.links[id] {
			links = append(links, link)

			if uint64(len(links)) == batchSize {
				return links, nil
			}
		}
	}

	return links, nil
}

func (r *LinksInMemoryService) TouchLink(_ context.Context, _ int64) error {
	return nil
}

func (r *LinksInMemoryService) UpdateLinkActivity(_ context.Context, _ int64, _ bool) error {
	return nil
}

// GetChatsIDs returns all the registered chat IDs.
func (r *LinksInMemoryService) GetChatsIDs(_ context.Context) ([]int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids := make([]int64, 0, len(r.links))
	for id := range r.links {
		ids = append(ids, id)
	}

	return ids, nil
}
