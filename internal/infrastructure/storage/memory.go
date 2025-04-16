package storage

import (
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
func (r *LinksInMemoryService) AddChat(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.links[id]; exists {
		return scrapperapi.ErrChatAlreadyExists
	}

	r.links[id] = make(map[string]models.Link)

	return nil
}

// DeleteChat deletes all links for the provided chat.
func (r *LinksInMemoryService) DeleteChat(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.links[id]; !exists {
		return scrapperapi.ErrChatNotFound
	}

	delete(r.links, id)

	return nil
}

// AddLink adds new tracking link.
func (r *LinksInMemoryService) AddLink(id int64, link models.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, exists := r.links[id]
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

// GetLinks retrieves links attached to the chat id.
func (r *LinksInMemoryService) GetLinks(id int64) (links []models.Link, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, exists := r.links[id]
	if !exists {
		return nil, scrapperapi.ErrChatNotFound
	}

	links = make([]models.Link, 0, len(entries))

	for _, link := range entries {
		links = append(links, link)
	}

	return links, nil
}

// DeleteLink deletes link attached to the chat id.
func (r *LinksInMemoryService) DeleteLink(id int64, url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.links[id]
	if !exists {
		return scrapperapi.ErrChatNotFound
	}

	entries := r.links[id]

	if _, exists := entries[url]; !exists {
		return scrapperapi.ErrLinkNotFound
	}

	delete(entries, url)

	return nil
}

// GetChatsIDs returns all the registered chat IDs.
func (r *LinksInMemoryService) GetChatIDs() ([]int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids := make([]int64, 0, len(r.links))
	for id := range r.links {
		ids = append(ids, id)
	}

	return ids, nil
}
