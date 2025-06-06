package commands

import (
	"context"
	"fmt"
	"net/http"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

type Untrack struct {
	Traits   *models.Traits
	Pipeline []*models.Stage
	Link     sclient.RemoveLinkRequest
	Client   Client
	Cache    Cache
}

func NewUntrack(chatID int64, client Client, cache Cache) *Untrack {
	return &Untrack{
		Traits:   models.NewTraits(UntrackSpan, chatID, CommandUntrack),
		Pipeline: createUntrackStages(),
		Client:   client,
		Cache:    cache,
	}
}

func (c *Untrack) Validate(input string) error {
	if err := c.Pipeline[c.Traits.Stage].Validate(input); err != nil {
		c.Traits.Malformed = true
		return err
	}

	c.Traits.HandleUntrack(input, &c.Link)

	return nil
}

func (c *Untrack) Stage() (string, bool) {
	if !c.Traits.Malformed {
		return c.Pipeline[c.Traits.Stage].Prompt, false
	}

	return c.Pipeline[c.Traits.Stage].Manual, false
}

func (c *Untrack) Done() bool {
	return c.Traits.Stage == c.Traits.Span
}

func (c *Untrack) Request(ctx context.Context) (string, error) {
	params := &sclient.DeleteLinksParams{TgChatId: c.Traits.ChatID}

	body := sclient.RemoveLinkRequest{Link: c.Link.Link}

	resp, err := c.Client.DeleteLinks(ctx, params, body)
	if err != nil {
		return FailedUntrack, fmt.Errorf("command untrack: failed to delete link: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return LinkNotYetTracked, nil
	}

	if resp.StatusCode != http.StatusOK {
		return FailedUntrack, fmt.Errorf("command untrack: client response code: %d", resp.StatusCode)
	}

	if err = c.Cache.Delete(ctx, c.Traits.ChatID, body); err != nil {
		return SuccessfulTrack, fmt.Errorf("command untrack: failed to delete link from cache: %w", err)
	}

	return SuccessfulUntrack, nil
}

func (c *Untrack) Name() string {
	return c.Traits.Name
}

func createUntrackStages() []*models.Stage {
	return []*models.Stage{
		models.NewStage(UntrackRequest, LinkManual, ValidateLink),
	}
}

const (
	CommandUntrack    = "untrack"
	UntrackSpan       = 1
	UntrackRequest    = "✨ Please, enter the link you want to untrack! (press /cancel to quit)"
	FailedUntrack     = "💥 Failed to untrack provided link."
	LinkNotYetTracked = "⚡️ This link is not yet being tracked!"
	SuccessfulUntrack = "✨ This link is no longer being tracked!"
)
