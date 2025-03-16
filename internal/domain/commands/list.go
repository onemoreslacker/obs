package commands

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
	scrcl "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/scrapper"
)

type CommandList struct {
	traits *entities.Traits

	pipeline []*entities.Stage
	tags     []string
	filters  []string

	scrapperClient scrcl.ClientInterface
}

func NewCommandList(
	chatID int64,
	client scrcl.ClientInterface,
	cfg *config.Config,
) *CommandList {
	return &CommandList{
		traits: entities.NewTraits(
			cfg.Meta.Spans.List,
			chatID,
			cfg.Meta.Commands.List,
		),
		pipeline:       createListStages(cfg),
		scrapperClient: client,
	}
}

func (c *CommandList) Validate(input string) error {
	if err := c.pipeline[c.traits.Stage].Validate(input); err != nil {
		c.traits.Malformed = true
		return err
	}

	c.traits.Malformed = false

	switch c.traits.Stage {
	case 0:
		if strings.ToLower(strings.TrimSpace(input)) == "no" {
			c.traits.Stage++
			c.tags = []string{}
		}
	case 1:
		c.tags = constructTags(input)
	case 2:
		if strings.ToLower(strings.TrimSpace(input)) == "no" {
			c.traits.Stage++
			c.filters = []string{}
		}
	case 3:
		filters := constructFilters(input)
		c.filters = filters
	}

	c.traits.Stage++

	return nil
}

func (c *CommandList) Stage() (string, bool) {
	keyboard := c.traits.Stage == 0 || c.traits.Stage == 2

	if !c.traits.Malformed {
		return c.pipeline[c.traits.Stage].Prompt, keyboard
	}

	return c.pipeline[c.traits.Stage].Manual, keyboard
}

func (c *CommandList) Done() bool {
	return c.traits.Stage == c.traits.Span
}

func (c *CommandList) Request() (any, error) {
	params := &scrcl.GetLinksParams{
		TgChatId: c.traits.ChatID,
	}

	resp, err := c.scrapperClient.GetLinks(context.Background(), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrLinksResponseFailed
	}

	var list scrcl.ListLinksResponse

	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	sievedLinks := make([]entities.Link, 0, *list.Size)

	for _, link := range *list.Links {
		if matchTags(*link.Tags, c.tags) &&
			matchFilters(*link.Filters, c.filters) {
			sievedLinks = append(sievedLinks, entities.Link(link))
		}
	}

	if len(sievedLinks) == 0 {
		return nil, ErrEmptyList
	}

	return sievedLinks, nil
}

func (c *CommandList) Name() string {
	return c.traits.Name
}

func createListStages(cfg *config.Config) []*entities.Stage {
	return []*entities.Stage{
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
