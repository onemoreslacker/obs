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
	GitHubHost                = "api.github.com"
	GitHubBasePath            = "repos"
	GitHubPRSuffix            = "pulls"
	GitHubIssueSuffix         = "issues"
	StackOverflowHost         = "api.stackexchange.com"
	StackOverflowAnswersPath  = "answers"
	StackOverflowCommentsPath = "comments"
)
