package external

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"time"
)

type GitHubRepository struct {
	LastUpdated time.Time `json:"updated_at"`
}

func (c *Client) GetGitHubRepository(link string) (GitHubRepository, error) {
	apiURL, err := c.buildGitHubAPIURL(link)
	if err != nil {
		return GitHubRepository{}, nil
	}

	req, err := http.NewRequest(http.MethodGet, apiURL, http.NoBody)
	if err != nil {
		return GitHubRepository{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GitHubRepository{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GitHubRepository{}, ErrRequestFailed
	}

	var repo GitHubRepository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return GitHubRepository{}, err
	}

	return repo, nil
}

func (c *Client) buildGitHubAPIURL(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	u.Host = c.GitHubHost
	u.Path = path.Join(c.GitHubBasePath, u.Path)

	return u.String(), nil
}
