package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

type StackOverflowClient struct {
	httpClient *http.Client
}

func NewStackOverflowClient() *StackOverflowClient {
	return &StackOverflowClient{
		httpClient: &http.Client{},
	}
}

type StackOverflowUpdate struct {
	Type  string
	Owner struct {
		Username string `json:"display_name"`
	} `json:"owner"`
	CreatedAt int64  `json:"creation_date"`
	Body      string `json:"body"`
}

type StackOverflowUpdates struct {
	Items []StackOverflowUpdate `json:"items"`
}

func (c *StackOverflowClient) RetrieveUpdates(ctx context.Context, link string) ([]models.Update, error) {
	answersURL, err := buildStackOverflowAPIURL(link, "answers")
	if err != nil {
		return nil, err
	}

	answers, err := c.fetchStackOverflowUpdates(ctx, answersURL)
	if err != nil {
		return nil, err
	}

	commentsURL, err := buildStackOverflowAPIURL(link, "comments")
	if err != nil {
		return nil, err
	}

	comments, err := c.fetchStackOverflowUpdates(ctx, commentsURL)
	if err != nil {
		return nil, err
	}

	updates := make([]models.Update, len(answers.Items)+len(comments.Items))

	for _, answer := range answers.Items {
		updates = append(updates, models.NewUpdate(
			"answer",
			time.Unix(answer.CreatedAt, 0).Format(time.RFC1123),
			answer.Owner.Username,
			answer.Body,
		))
	}

	for _, comment := range comments.Items {
		updates = append(updates, models.NewUpdate(
			"comment",
			comment.Owner.Username,
			time.Unix(comment.CreatedAt, 0).Format(time.RFC1123),
			comment.Body,
		))
	}

	return updates, nil
}

func (c *StackOverflowClient) fetchStackOverflowUpdates(ctx context.Context, apiURL string) (StackOverflowUpdates, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
	if err != nil {
		return StackOverflowUpdates{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return StackOverflowUpdates{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return StackOverflowUpdates{}, ErrRequestFailed
	}

	var answers StackOverflowUpdates

	if err := json.NewDecoder(resp.Body).Decode(&answers.Items); err != nil {
		return StackOverflowUpdates{}, err
	}

	return answers, nil
}

func buildStackOverflowAPIURL(link, basePath string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	parts := strings.Split(u.Path, "/")
	parts = parts[:len(parts)-1]

	cut := path.Join(strings.Join(parts, "/"), basePath)
	u.Path = path.Join("api.stackexchange.com", cut)

	query := u.Query()
	query.Set("order", "desc")
	query.Set("sort", "activity")
	query.Set("site", "stackoverflow")
	query.Set("filter", "withbody")
	u.RawQuery = query.Encode()

	return u.String(), nil
}
