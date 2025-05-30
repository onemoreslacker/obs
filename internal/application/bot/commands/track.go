package commands

import (
	"context"
	"net/http"
	"time"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

type Poster interface {
	PostLinks(ctx context.Context, params *sclient.PostLinksParams, body sclient.PostLinksJSONRequestBody,
		reqEditors ...sclient.RequestEditorFn) (*http.Response, error)
}

type Track struct {
	Traits   *models.Traits
	Pipeline []*models.Stage
	Link     sclient.AddLinkRequest
	Client   Poster
}

func NewTrack(chatID int64, client Poster) *Track {
	return &Track{
		Traits:   models.NewTraits(TrackSpan, chatID, CommandTrack),
		Pipeline: createTrackStages(),
		Link:     sclient.AddLinkRequest{},
		Client:   client,
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

func (c *Track) Request() string {
	params := &sclient.PostLinksParams{
		TgChatId: c.Traits.ChatID,
	}

	body := sclient.AddLinkRequest{
		Link:    c.Link.Link,
		Tags:    c.Link.Tags,
		Filters: c.Link.Filters,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.Client.PostLinks(ctx, params, body)
	if err != nil {
		return FailedTrack
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return LinkAlreadyTracked
	}

	if resp.StatusCode != http.StatusOK {
		return FailedTrack
	}

	return SuccessfulTrack
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
	TrackRequest       = "✨ Please, enter the Link you want to track! (press /cancel to quit)"
	FailedTrack        = "💥 Failed to track Link!"
	LinkAlreadyTracked = "⚡️ This Link is already being tracked!"
	SuccessfulTrack    = "✨ This Link is now being tracked!"
)
