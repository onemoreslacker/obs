package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

type Track struct {
	Traits   *models.Traits
	Pipeline []*models.Stage
	Link     sclient.AddLinkRequest
	Client   Client
	Cache    Cache
}

func NewTrack(chatID int64, client Client, cache Cache) *Track {
	return &Track{
		Traits:   models.NewTraits(TrackSpan, chatID, CommandTrack),
		Pipeline: createTrackStages(),
		Link:     sclient.AddLinkRequest{},
		Client:   client,
		Cache:    cache,
	}
}

func (c *Track) Validate(input string) error {
	if err := c.Pipeline[c.Traits.Stage].Validate(input); err != nil {
		c.Traits.Malformed = true
		return err
	}

	c.Traits.HandleTrack(input, &c.Link)

	return nil
}

func (c *Track) Stage() (string, bool) {
	keyboard := c.Traits.Stage == 1 || c.Traits.Stage == 3

	if !c.Traits.Malformed {
		return c.Pipeline[c.Traits.Stage].Prompt, keyboard
	}

	return c.Pipeline[c.Traits.Stage].Manual, keyboard
}

func (c *Track) Done() bool {
	return c.Traits.Stage == c.Traits.Span
}

func (c *Track) Request(ctx context.Context) (string, error) {
	params := &sclient.PostLinksParams{
		TgChatId: c.Traits.ChatID,
	}

	body := sclient.AddLinkRequest{
		Link:    c.Link.Link,
		Tags:    c.Link.Tags,
		Filters: c.Link.Filters,
	}

	resp, err := c.Client.PostLinks(ctx, params, body)
	if err != nil {
		return FailedTrack, fmt.Errorf("command track: failed to post link: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return LinkAlreadyTracked, nil
	}

	if resp.StatusCode != http.StatusOK {
		return FailedTrack, fmt.Errorf("command list: client response code: %d", resp.StatusCode)
	}

	var link sclient.LinkResponse
	if err = json.NewDecoder(resp.Body).Decode(&link); err != nil {
		return SuccessfulTrack, fmt.Errorf("command track: failed to deserialize link response: %w", err)
	}

	if err = c.Cache.Add(ctx, c.Traits.ChatID, link); err != nil {
		return SuccessfulTrack, fmt.Errorf("command track: failed to add new link to cache: %w", err)
	}

	return SuccessfulTrack, nil
}

func (c *Track) Name() string {
	return c.Traits.Name
}

func createTrackStages() []*models.Stage {
	return []*models.Stage{
		models.NewStage(TrackRequest, LinkManual, ValidateLink),
		models.NewStage(TagsAck, TagsManual, ValidateAck),
		models.NewStage(TagsRequest, TagsManual, ValidateTags),
		models.NewStage(FiltersAck, AcksManual, ValidateAck),
		models.NewStage(FiltersRequest, FiltersManual, ValidateFilters),
	}
}

const (
	CommandTrack       = "track"
	TrackSpan          = 5
	TrackRequest       = "✨ Please, enter the link you want to track! (press /cancel to quit)"
	FailedTrack        = "💥 Failed to track link!"
	LinkAlreadyTracked = "⚡️ This link is already being tracked!"
	SuccessfulTrack    = "✨ This link is now being tracked!"
)
