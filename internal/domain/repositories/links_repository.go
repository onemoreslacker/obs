package repositories

import "github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"

type LinksRepository interface {
	AddChat(id int64) error
	DeleteChat(id int64) error
	AddLink(id int64, link entities.Link) error
	GetLinks(id int64) ([]entities.Link, error)
	DeleteLink(id int64, url string) (entities.Link, error)
	GetChatIDs() []int64
}
