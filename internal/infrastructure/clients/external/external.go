package external

import (
	"net/http"
)

type Client struct {
	httpClient *http.Client
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

const (
	GitHubHost            = "api.github.com"
	GitHubBasePath        = "repos"
	StackOverflowHost     = "api.stackexchange.com"
	StackOverflowBasePath = "answers"
)
