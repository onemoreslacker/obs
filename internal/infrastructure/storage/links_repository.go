package storage

import (
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
)

// LinksRepository stores links tracked by user and guarded by mutex.
type LinksRepository struct {
	links map[int64]map[string]entities.Link
	mu    sync.Mutex
}

// NewLinksRepository implements a new LinksRepository entity.
func NewLinksRepository() *LinksRepository {
	return &LinksRepository{
		links: make(map[int64]map[string]entities.Link),
	}
}

// AddChat creates a map of links for a new chat.
func (r *LinksRepository) AddChat(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.links[id]; exists {
		return ErrChatAlreadyExists
	}

	r.links[id] = make(map[string]entities.Link)

	return nil
}

// DeleteChat deletes all links for the provided chat.
func (r *LinksRepository) DeleteChat(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.links[id]; !exists {
		return ErrChatNotFound
	}

	delete(r.links, id)

	return nil
}

// AddLink adds new tracking link.
func (r *LinksRepository) AddLink(id int64, link entities.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, exists := r.links[id]
	if !exists {
		return ErrChatNotFound
	}

	if _, exists := entries[*link.Url]; exists {
		return ErrLinkAlreadyExists
	}

	entries[*link.Url] = link

	return nil
}

// GetLinks retrieves links attached to the chat id.
func (r *LinksRepository) GetLinks(id int64) (links []entities.Link, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, exists := r.links[id]
	if !exists {
		return nil, ErrChatNotFound
	}

	links = make([]entities.Link, 0, len(entries))

	for _, link := range entries {
		links = append(links, link)
	}

	return links, nil
}

// DeleteLink deletes link attached to the chat id.
func (r *LinksRepository) DeleteLink(id int64, url string) (link entities.Link, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.links[id]
	if !exists {
		return entities.Link{}, ErrChatNotFound
	}

	entries := r.links[id]

	if _, exists := entries[url]; !exists {
		return entities.Link{}, ErrLinkNotFound
	}

	link = entries[url]

	delete(entries, url)

	return link, nil
}

// GetChatsIDs returns all the registered chat IDs.
func (r *LinksRepository) GetChatIDs() []int64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids := make([]int64, 0, len(r.links))
	for id := range r.links {
		ids = append(ids, id)
	}

	return ids
}
