package clients

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/fetcher"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"resty.dev/v3"
)

type GitHubClient struct {
	client *resty.Client
}

func NewGithubClient(cfg *config.Config) *GitHubClient {
	cb := resty.NewCircuitBreaker().
		SetTimeout(cfg.CircuitBreakerPolicy.Timeout).
		SetFailureThreshold(cfg.CircuitBreakerPolicy.FailureThreshold).
		SetSuccessThreshold(cfg.CircuitBreakerPolicy.MaxRequests)

	client := resty.New().
		SetBaseURL("https://api.github.com/repos/").
		SetAuthScheme("Bearer").
		SetAuthToken(cfg.Secrets.GitHubToken).
		SetTimeout(cfg.TimeoutPolicy.ClientOverall).
		SetRetryCount(int(cfg.RetryPolicy.Attempts)).
		SetRetryWaitTime(cfg.RetryPolicy.Delay).
		AddRetryConditions(func(res *resty.Response, err error) bool {
			return !slices.Contains(cfg.RetryPolicy.StatusCodes, res.StatusCode())
		}).
		SetCircuitBreaker(cb)

	return &GitHubClient{
		client: client,
	}
}

type GitHubUpdate struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	User  struct {
		Login string `json:"login"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
}

func (g *GitHubClient) RetrieveUpdates(ctx context.Context, link string) ([]models.Update, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("github client: failed to parse link")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 {
		return nil, fetcher.ErrInvalidRepoPath
	}

	g.client.SetContext(ctx).
		SetPathParams(map[string]string{
			"owner": parts[0],
			"repo":  parts[1],
		}).
		SetQueryParams(map[string]string{
			"sort":      "updated",
			"direction": "desc",
		})

	var pulls []GitHubUpdate
	if _, err := g.client.R().
		SetResult(&pulls).
		Get("{owner}/{repo}/pulls"); err != nil {
		return nil, fmt.Errorf("github client: failed to fetch issues updates")
	}

	var issues []GitHubUpdate
	if _, err := g.client.R().
		SetResult(&issues).
		Get("{owner}/{repo}/issues"); err != nil {
		return nil, fmt.Errorf("github client: failed to fetch issues updates")
	}

	updates := make([]models.Update, 0, len(pulls)+len(issues))

	for _, pull := range pulls {
		updates = append(updates, models.NewUpdate(
			pull.Title,
			pull.CreatedAt,
			pull.User.Login,
			pull.Body,
		))
	}

	for _, issue := range issues {
		updates = append(updates, models.NewUpdate(
			issue.Title,
			issue.CreatedAt,
			issue.User.Login,
			issue.Body,
		))
	}

	return updates, nil
}

func (g *GitHubClient) Close() error {
	if err := g.client.Close(); err != nil {
		return fmt.Errorf("github client: failed to cleanup resources: %w", err)
	}

	return nil
}
