package commands

import (
	"context"
	"log"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	scrcl "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/scrapper"
)

type CommandUntrack struct {
	traits *models.Traits

	pipeline []*models.Stage
	link     models.Link

	scrapperClient scrcl.ClientInterface
}

func NewCommandUntrack(
	chatID int64,
	client scrcl.ClientInterface,
) *CommandUntrack {
	return &CommandUntrack{
		traits: models.NewTraits(
			UntrackSpan,
			chatID,
			Untrack,
		),
		pipeline:       createUntrackStages(),
		scrapperClient: client,
	}
}

func (c *CommandUntrack) Validate(input string) error {
	if err := c.pipeline[c.traits.Stage].Validate(input); err != nil {
		c.traits.Malformed = true
		return err
	}

	c.traits.HandleUntrack(input, &c.link)

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

func (c *CommandUntrack) Request() string {
	params := &scrcl.DeleteLinksParams{
		TgChatId: c.traits.ChatID,
	}

	body := scrcl.DeleteLinksJSONRequestBody{
		Link: c.link.Url,
	}

	resp, err := c.scrapperClient.DeleteLinks(
		context.Background(), params, body)
	if err != nil {
		return FailedUntrack
	}
	defer resp.Body.Close()

	log.Println(resp.StatusCode)

	if resp.StatusCode == http.StatusBadRequest {
		return FailedUntrack
	}

	if resp.StatusCode == http.StatusConflict {
		return LinkNotYetTracked
	}

	return SuccessfulUntrack
}

func (c *CommandUntrack) Name() string {
	return c.traits.Name
}

func createUntrackStages() []*models.Stage {
	return []*models.Stage{
		models.NewStage(
			UntrackRequest,
			LinkManual,
			ValidateLink,
		),
	}
}

const (
	Untrack           = "untrack"
	UntrackSpan       = 1
	UntrackRequest    = "✨ Please, enter the link you want to untrack! (press /cancel to quit)"
	FailedUntrack     = "💥 Failed to untrack provided link."
	LinkNotYetTracked = "⚡️ This link is not yet being tracked!"
	SuccessfulUntrack = "✨ This link is no longer being tracked!"
)
