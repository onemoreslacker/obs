package clients

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/fetcher"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"resty.dev/v3"
)

type StackOverflowClient struct {
	client *resty.Client
}

func NewStackOverflowClient(cfg *config.Config) *StackOverflowClient {
	cb := resty.NewCircuitBreaker().
		SetTimeout(5 * time.Second).
		SetFailureThreshold(3).
		SetSuccessThreshold(1)

	client := resty.New().
		SetBaseURL("https://api.stackexchange.com/2.3/questions/").
		SetHeader("X-API-Access", cfg.Secrets.StackOverflowToken).
		SetTimeout(cfg.TimeoutPolicy.ClientOverall).
		SetRetryCount(int(cfg.RetryPolicy.Attempts)).
		SetRetryWaitTime(cfg.RetryPolicy.Delay).
		AddRetryConditions(func(res *resty.Response, err error) bool {
			return !slices.Contains(cfg.RetryPolicy.StatusCodes, res.StatusCode())
		}).
		SetCircuitBreaker(cb)

	return &StackOverflowClient{
		client: client,
	}
}

type StackOverflowUpdate struct {
	Body  string `json:"body"`
	Owner struct {
		Username string `json:"display_name"`
	} `json:"owner"`
	CreatedAt int64 `json:"creation_date"`
}

type StackOverflowUpdates struct {
	Items []StackOverflowUpdate `json:"items"`
}

func (s *StackOverflowClient) RetrieveUpdates(ctx context.Context, link string) ([]models.Update, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("stackoverflow client: failed to parse link")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return nil, fetcher.ErrInvalidRepoPath
	}

	questionID := parts[1]

	s.client.SetContext(ctx).
		SetPathParams(map[string]string{
			"questionID": questionID,
		}).
		SetQueryParams(map[string]string{
			"order":  "desc",
			"sort":   "activity",
			"site":   "stackoverflow",
			"filter": "withbody",
		})

	var answers StackOverflowUpdates
	if _, err := s.client.R().
		SetResult(&answers).
		Get("{questionID}/answers"); err != nil {
		return nil, fmt.Errorf("stackoverflow client: failed to fetch answers updates")
	}

	var comments StackOverflowUpdates
	if _, err := s.client.R().
		SetResult(&comments).
		Get("{questionID}/comments"); err != nil {
		return nil, fmt.Errorf("stackoverflow client: failed to fetch comments updates")
	}

	updates := make([]models.Update, 0, len(answers.Items)+len(comments.Items))

	for _, answer := range answers.Items {
		updates = append(updates, models.NewUpdate(
			"answer",
			time.Unix(answer.CreatedAt, 0).Format(time.RFC3339),
			answer.Owner.Username,
			answer.Body,
		))
	}

	for _, comment := range comments.Items {
		updates = append(updates, models.NewUpdate(
			"comment",
			time.Unix(comment.CreatedAt, 0).Format(time.RFC3339),
			comment.Owner.Username,
			comment.Body,
		))
	}

	return updates, nil
}

func (s *StackOverflowClient) Close() error {
	if err := s.client.Close(); err != nil {
		return fmt.Errorf("stackoverflow client: failed to cleanup resources: %w", err)
	}

	return nil
}
