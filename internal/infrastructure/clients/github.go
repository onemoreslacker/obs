package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

type GitHubClient struct {
	httpClient *http.Client
}

func NewGithubClient() *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{},
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

type GitHubUpdates struct {
	Items []GitHubUpdate `json:"items"`
}

func (c *GitHubClient) RetrieveUpdates(ctx context.Context, link string) ([]models.Update, error) {
	prURL, err := buildGitHubAPIURL(link, "pulls")
	if err != nil {
		return nil, err
	}

	pulls, err := c.fetchGitHubUpdates(ctx, prURL)
	if err != nil {
		return nil, err
	}

	issuesURL, err := buildGitHubAPIURL(link, "issues")
	if err != nil {
		return nil, err
	}

	issues, err := c.fetchGitHubUpdates(ctx, issuesURL)
	if err != nil {
		return nil, err
	}

	updates := make([]models.Update, 0, len(pulls.Items)+len(issues.Items))

	for _, pull := range pulls.Items {
		updates = append(updates, models.NewUpdate(
			pull.Title,
			pull.CreatedAt,
			pull.User.Login,
			pull.Body,
		))
	}

	for _, issue := range issues.Items {
		updates = append(updates, models.NewUpdate(
			issue.Title,
			issue.CreatedAt,
			issue.User.Login,
			issue.Body,
		))
	}

	return updates, nil
}

func (c *GitHubClient) fetchGitHubUpdates(ctx context.Context, apiURL string) (GitHubUpdates, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
	if err != nil {
		return GitHubUpdates{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GitHubUpdates{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GitHubUpdates{}, ErrRequestFailed
	}

	var updates GitHubUpdates

	if err := json.NewDecoder(resp.Body).Decode(&updates.Items); err != nil {
		return GitHubUpdates{}, err
	}

	return updates, nil
}

func buildGitHubAPIURL(link, suffix string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	u.Host = "api.github.com"
	u.Path = path.Join("repos", u.Path, suffix)

	query := u.Query()
	query.Set("sort", "updated")
	query.Set("direction", "desc")
	u.RawQuery = query.Encode()

	return u.String(), nil
}
