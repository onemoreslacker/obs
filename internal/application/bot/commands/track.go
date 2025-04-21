package commands

import (
	"context"
	"net/http"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	scrcl "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/scrapper"
)

type CommandTrack struct {
	traits *models.Traits

	pipeline []*models.Stage
	link     models.Link

	scrapperClient scrcl.ClientInterface
}

func NewCommandTrack(
	chatID int64,
	client scrcl.ClientInterface,
) *CommandTrack {
	return &CommandTrack{
		traits: models.NewTraits(
			TrackSpan,
			chatID,
			Track,
		),
		pipeline:       createTrackStages(),
		link:           models.Link{Id: &chatID},
		scrapperClient: client,
	}
}

func (c *CommandTrack) Validate(input string) error {
	if err := c.pipeline[c.traits.Stage].Validate(input); err != nil {
		c.traits.Malformed = true
		return err
	}

	c.traits.HandleTrack(input, &c.link)

	return nil
}

func (c *CommandTrack) Stage() (string, bool) {
	keyboard := c.traits.Stage == 1 || c.traits.Stage == 3

	if !c.traits.Malformed {
		return c.pipeline[c.traits.Stage].Prompt, keyboard
	}

	return c.pipeline[c.traits.Stage].Manual, keyboard
}

func (c *CommandTrack) Done() bool {
	return c.traits.Stage == c.traits.Span
}

func (c *CommandTrack) Request() string {
	params := &scrcl.PostLinksParams{
		TgChatId: c.traits.ChatID,
	}

	body := scrcl.PostLinksJSONRequestBody{
		Link:    c.link.Url,
		Tags:    c.link.Tags,
		Filters: c.link.Filters,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.scrapperClient.PostLinks(ctx, params, body)
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

func (c *CommandTrack) Name() string {
	return c.traits.Name
}

func createTrackStages() []*models.Stage {
	return []*models.Stage{
		models.NewStage(
			TrackRequest,
			LinkManual,
			ValidateLink,
		),
		models.NewStage(
			TagsAck,
			TagsManual,
			validateAck,
		),
		models.NewStage(
			TagsRequest,
			TagsManual,
			validateTags,
		),
		models.NewStage(
			FiltersAck,
			AcksManual,
			validateAck,
		),
		models.NewStage(
			FiltersRequest,
			FiltersManual,
			validateFilters,
		),
	}
}

const (
	Track              = "track"
	TrackSpan          = 5
	TrackRequest       = "✨ Please, enter the link you want to track! (press /cancel to quit)"
	FailedTrack        = "💥 Failed to track link!"
	LinkAlreadyTracked = "⚡️ This link is already being tracked!"
	SuccessfulTrack    = "✨ This link is now being tracked!"
)
