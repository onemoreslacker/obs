package clients

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

type Client interface {
	RetrieveUpdates(ctx context.Context, link string) ([]models.Update, error)
}

func New(source string) Client {
	switch source {
	case config.GitHub:
		return NewGithubClient()
	case config.StackOverflow:
		return NewStackOverflowClient()
	}

	return nil
}
