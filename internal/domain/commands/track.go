package commands

import (
	"context"
	"net/http"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
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
	cfg *config.Config,
) *CommandTrack {
	return &CommandTrack{
		traits: entities.NewTraits(
			cfg.Meta.Spans.Track,
			chatID,
			cfg.Meta.Commands.Track,
		),
		pipeline:       createTrackStages(cfg),
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

func (c *CommandTrack) Request() (any, error) {
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
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrFailedToTrack
	}

	return nil, nil
}

func (c *CommandTrack) Name() string {
	return c.traits.Name
}

func createTrackStages(cfg *config.Config) []*entities.Stage {
	return []*entities.Stage{
		entities.NewStage(
			cfg.Meta.Replies.Track,
			cfg.Meta.Manuals.Link,
			ValidateLink,
		),
		entities.NewStage(
			cfg.Meta.Replies.TagsAck,
			cfg.Meta.Manuals.Acks,
			validateAck,
		),
		entities.NewStage(
			cfg.Meta.Replies.Tags,
			cfg.Meta.Manuals.Tags,
			validateTags,
		),
		entities.NewStage(
			cfg.Meta.Replies.FiltersAck,
			cfg.Meta.Manuals.Acks,
			validateAck,
		),
		entities.NewStage(
			cfg.Meta.Replies.Filters,
			cfg.Meta.Manuals.Filters,
			validateFilters,
		),
	}
}
