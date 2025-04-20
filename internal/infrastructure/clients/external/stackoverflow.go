package external

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

// StackOverflowUpdates represents updates via the specified link (either answer or comment).
type StackOverflowUpdates struct {
	Items []models.StackOverflowUpdate `json:"items"`
}

// RetrieveStackOverflowUpdates returns updates from StackOverflow associated with answers and comments.
func (c *Client) RetrieveStackOverflowUpdates(link string) ([]models.StackOverflowUpdate, error) {
	answersURL, err := buildStackOverflowAPIURL(link, StackOverflowAnswersPath)
	if err != nil {
		return nil, err
	}

	answers, err := c.fetchStackOverflowUpdates(answersURL)
	if err != nil {
		return nil, err
	}

	commentsURL, err := buildStackOverflowAPIURL(link, StackOverflowCommentsPath)
	if err != nil {
		return nil, err
	}

	comments, err := c.fetchStackOverflowUpdates(commentsURL)
	if err != nil {
		return nil, err
	}

	updates := make([]models.StackOverflowUpdate, 0, len(answers.Items)+len(comments.Items))

	for _, answer := range answers.Items {
		answer.Type = "answer"
		updates = append(updates, answer)
	}

	for _, comment := range comments.Items {
		comment.Type = "comment"
		updates = append(updates, comment)
	}

	return updates, nil
}

// fetchStackOverflowUpdates fetches updates from StackOverflow, whether it is answers or comments.
func (c *Client) fetchStackOverflowUpdates(apiURL string) (StackOverflowUpdates, error) {
	req, err := http.NewRequest(http.MethodGet, apiURL, http.NoBody)
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

// buildStackOverflowAPIURL builds url according to provided link and basePath.
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
	u.Path = path.Join(StackOverflowHost, cut)

	query := u.Query()
	query.Set("order", "desc")
	query.Set("sort", "activity")
	query.Set("site", "stackoverflow")
	query.Set("filter", "withbody")
	u.RawQuery = query.Encode()

	return u.String(), nil
}
