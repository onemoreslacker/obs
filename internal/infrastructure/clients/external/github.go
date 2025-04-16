package external

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

type GitHubUpdates struct {
	Items []models.GitHubUpdate `json:"items"`
}

func (c *Client) RetrieveGitHubUpdates(link string) ([]models.GitHubUpdate, error) {
	prURL, err := buildGitHubAPIURL(link, GitHubPRSuffix)
	if err != nil {
		return nil, err
	}

	pulls, err := c.fetchGitHubUpdates(prURL)
	if err != nil {
		return nil, err
	}

	issuesURL, err := buildGitHubAPIURL(link, GitHubIssueSuffix)
	if err != nil {
		return nil, err
	}

	issues, err := c.fetchGitHubUpdates(issuesURL)
	if err != nil {
		return nil, err
	}

	updates := make([]models.GitHubUpdate, 0, len(pulls.Items)+len(issues.Items))

	updates = append(updates, pulls.Items...)
	updates = append(updates, issues.Items...)

	return updates, nil
}

func (c *Client) fetchGitHubUpdates(apiURL string) (GitHubUpdates, error) {
	req, err := http.NewRequest(http.MethodGet, apiURL, http.NoBody)
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

	u.Host = GitHubHost
	u.Path = path.Join(GitHubBasePath, u.Path, suffix)

	query := u.Query()
	query.Set("sort", "created")
	query.Set("direction", "desc")
	u.RawQuery = query.Encode()

	return u.String(), nil
}
