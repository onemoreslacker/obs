package commands

import (
	"context"
	"net/http"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
	scrcl "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/scrapper"
)

type CommandTrack struct {
	traits *entities.Traits

	pipeline []*entities.Stage
	link     entities.Link

	scrapperClient scrcl.ClientInterface
}

func NewCommandTrack(
	chatID int64,
	client scrcl.ClientInterface,
) *CommandTrack {
	return &CommandTrack{
		traits: entities.NewTraits(
			TrackSpan,
			chatID,
			Track,
		),
		pipeline:       createTrackStages(),
		link:           entities.Link{Id: &chatID},
		scrapperClient: client,
	}
}

func (c *CommandTrack) Validate(input string) error {
	if err := c.pipeline[c.traits.Stage].Validate(input); err != nil {
		c.traits.Malformed = true
		return err
	}

	c.traits.Malformed = false

	switch c.traits.Stage {
	case 0:
		c.link.Url = &input
	case 1:
		if strings.ToLower(strings.TrimSpace(input)) == "no" {
			c.traits.Stage++
			c.link.Tags = &[]string{}
		}
	case 2:
		tags := constructTags(input)
		c.link.Tags = &tags
	case 3:
		if strings.ToLower(strings.TrimSpace(input)) == "no" {
			c.traits.Stage++
			c.link.Filters = &[]string{}
		}
	case 4:
		filters := constructFilters(input)
		c.link.Filters = &filters
	}

	c.traits.Stage++

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

	resp, err := c.scrapperClient.PostLinks(
		context.Background(), params, body)
	if err != nil {
		return FailedTrack
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return LinkAlreadyTracked
	}

	return SuccessfulTrack
}

func (c *CommandTrack) Name() string {
	return c.traits.Name
}

func createTrackStages() []*entities.Stage {
	return []*entities.Stage{
		entities.NewStage(
			TrackRequest,
			LinkManual,
			ValidateLink,
		),
		entities.NewStage(
			TagsAck,
			TagsManual,
			validateAck,
		),
		entities.NewStage(
			TagsRequest,
			TagsManual,
			validateTags,
		),
		entities.NewStage(
			FiltersAck,
			AcksManual,
			validateAck,
		),
		entities.NewStage(
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
