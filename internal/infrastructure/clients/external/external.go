package external

import (
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
)

type Client struct {
	httpClient            *http.Client
	GitHubHost            string
	GitHubBasePath        string
	StackOverflowHost     string
	StackOverflowBasePath string
}

func New(cfg *config.Config) *Client {
	return &Client{
		httpClient:            &http.Client{},
		GitHubHost:            cfg.Meta.Services.GitHubHost,
		GitHubBasePath:        cfg.Meta.Services.GitHubBasePath,
		StackOverflowHost:     cfg.Meta.Services.StackOverflowHost,
		StackOverflowBasePath: cfg.Meta.Services.StackOverflowBasePath,
	}
}
