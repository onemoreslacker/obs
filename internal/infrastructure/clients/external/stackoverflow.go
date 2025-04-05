package external

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type StackOverflowAnswers struct {
	Items []struct {
		LastActivityDate int64 `json:"last_activity_date"`
	} `json:"items"`
}

func (c *Client) GetStackOverflowAnswers(link string) (StackOverflowAnswers, error) {
	apiURL, err := buildStackOverflowAPIURL(link)
	if err != nil {
		return StackOverflowAnswers{}, err
	}

	req, err := http.NewRequest(http.MethodGet, apiURL, http.NoBody)
	if err != nil {
		return StackOverflowAnswers{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return StackOverflowAnswers{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return StackOverflowAnswers{}, ErrRequestFailed
	}

	var answers StackOverflowAnswers
	if err := json.NewDecoder(resp.Body).Decode(&answers); err != nil {
		return StackOverflowAnswers{}, err
	}

	return answers, nil
}

func buildStackOverflowAPIURL(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	parts := strings.Split(u.Path, "/")
	parts = parts[:len(parts)-1]

	cut := path.Join(strings.Join(parts, "/"), StackOverflowBasePath)
	u.Path = cut

	query := u.Query()
	query.Set("order", "desc")
	query.Set("sort", "activity")
	query.Set("site", "stackoverflow")
	u.RawQuery = query.Encode()

	return u.String(), nil
}
