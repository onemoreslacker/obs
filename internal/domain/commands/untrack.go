package commands

import (
	"context"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
	scrcl "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/scrapper"
)

type CommandUntrack struct {
	traits *entities.Traits

	pipeline []*entities.Stage
	url      entities.URL

	scrapperClient scrcl.ClientInterface
}

func NewCommandUntrack(
	chatID int64,
	client scrcl.ClientInterface,
	cfg *config.Config,
) *CommandUntrack {
	return &CommandUntrack{
		traits: entities.NewTraits(
			cfg.Meta.Spans.Untrack,
			chatID,
			cfg.Meta.Commands.Untrack,
		),
		pipeline:       createUntrackStages(cfg),
		url:            entities.URL{},
		scrapperClient: client,
	}
}

func (c *CommandUntrack) Validate(input string) error {
	if err := c.pipeline[c.traits.Stage].Validate(input); err != nil {
		c.traits.Malformed = true
		return err
	}

	c.traits.Malformed = false

	c.url.Link = &input

	c.traits.Stage++

	return nil
}

func (c *CommandUntrack) Stage() (string, bool) {
	keyboard := false

	if !c.traits.Malformed {
		return c.pipeline[c.traits.Stage].Prompt, keyboard
	}

	return c.pipeline[c.traits.Stage].Manual, keyboard
}

func (c *CommandUntrack) Done() bool {
	return c.traits.Stage == c.traits.Span
}

func (c *CommandUntrack) Request() (any, error) {
	params := &scrcl.DeleteLinksParams{
		TgChatId: c.traits.ChatID,
	}

	body := scrcl.DeleteLinksJSONRequestBody{
		Link: c.url.Link,
	}

	resp, err := c.scrapperClient.DeleteLinks(
		context.Background(), params, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrFailedToUntrack
	}

	return nil, nil
}

func (c *CommandUntrack) Name() string {
	return c.traits.Name
}

func createUntrackStages(cfg *config.Config) []*entities.Stage {
	return []*entities.Stage{
		entities.NewStage(
			cfg.Meta.Replies.Untrack,
			cfg.Meta.Manuals.Link,
			ValidateLink,
		),
	}
}
