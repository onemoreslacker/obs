package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
)

const (
	FiltersAck = "❔ Do you want to specify filters? (press /cancel to quit)"
	TagsAck    = "❔ Do you want to specify tags? (press /cancel to quit)"

	TagsRequest    = "✨ Please, enter link tags separated by space. (press /cancel to quit)"
	FiltersRequest = "✨ Please, enter link filters as filter:value. (press /cancel to quit)"

	LinkManual    = "💥 Invalid URL! Please enter a valid link (e.g. https://github.com/golang/go)"
	AcksManual    = "💥 Only yes/no are acceptable!"
	TagsManual    = "💥 Invalid tags! Use spaces to separate (e.g. 'work hobby')."
	FiltersManual = "💥 Invalid format! Use 'filter:value' (e.g. 'user:dummy')."
)

type Client interface {
	PostLinks(ctx context.Context, params *sclient.PostLinksParams, body sclient.PostLinksJSONRequestBody,
		reqEditors ...sclient.RequestEditorFn) (*http.Response, error)
	DeleteLinks(ctx context.Context, params *sclient.DeleteLinksParams, body sclient.DeleteLinksJSONRequestBody,
		reqEditors ...sclient.RequestEditorFn) (*http.Response, error)
	GetLinks(ctx context.Context, params *sclient.GetLinksParams,
		reqEditors ...sclient.RequestEditorFn) (*http.Response, error)
}

type Cache interface {
	Add(ctx context.Context, chatID int64, link sclient.LinkResponse) error
	Delete(ctx context.Context, chatID int64, req sclient.RemoveLinkRequest) error
	Get(ctx context.Context, chatID int64) (sclient.ListLinksResponse, error)
}

func ConstructListMessage(links []sclient.LinkResponse) string {
	var buf bytes.Buffer

	for i, link := range links {
		fmt.Fprintf(&buf, "%d. %s\n", i+1, link.Url)
	}

	return buf.String()
}

func MatchTags(got, desired []string) bool {
	if len(got) != len(desired) {
		return false
	}

	for _, tag := range desired {
		if !slices.Contains(got, tag) {
			return false
		}
	}

	return true
}

func MatchFilters(got, desired []string) bool {
	if len(got) != len(desired) {
		return false
	}

	for _, filter := range desired {
		if !slices.Contains(got, filter) {
			return false
		}
	}

	return true
}

func ValidateLink(link string) error {
	_, err := url.Parse(link)
	if err != nil {
		return ErrInvalidLinkFormat
	}

	if !((strings.Contains(link, config.StackOverflow) && strings.Contains(link, "questions")) ||
		strings.Contains(link, config.GitHub)) {
		return ErrInvalidLinkFormat
	}

	return nil
}

func ValidateAck(input string) error {
	if !slices.Contains([]string{"yes", "no"},
		strings.ToLower(strings.TrimSpace(input))) {
		return ErrInvalidAck
	}

	return nil
}

func ValidateTags(input string) error {
	if !(input == "" || len(strings.Fields(input)) > 0) {
		return ErrInvalidTagsFormat
	}

	return nil
}

func ValidateFilters(input string) error {
	filters := strings.Fields(input)

	for _, filter := range filters {
		pair := strings.Split(filter, ":")
		if len(pair) != 2 {
			return ErrInvalidFiltersFormat
		}
	}

	return nil
}

func (c *List) GetLinksWithCache(ctx context.Context) (sclient.ListLinksResponse, error) {
	cached, err := c.Cache.Get(ctx, c.Traits.ChatID)
	if err == nil {
		return cached, nil
	}

	params := &sclient.GetLinksParams{TgChatId: c.Traits.ChatID}
	resp, err := c.Client.GetLinks(ctx, params)
	if err != nil {
		return sclient.ListLinksResponse{}, fmt.Errorf("command list: failed to get links: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return sclient.ListLinksResponse{}, fmt.Errorf("command list: client response code: %d", resp.StatusCode)
	}

	var links sclient.ListLinksResponse
	if err = json.NewDecoder(resp.Body).Decode(&links); err != nil {
		return sclient.ListLinksResponse{}, fmt.Errorf("command list: failed to decode links response: %w", err)
	}

	for _, link := range links.Links {
		if err = c.Cache.Add(ctx, c.Traits.ChatID, link); err != nil {
			return links, fmt.Errorf("command list: failed to validate cache: %w", err)
		}
	}

	return links, nil
}
