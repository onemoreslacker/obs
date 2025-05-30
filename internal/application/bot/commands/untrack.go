package commands

import (
	"context"
	"net/http"
	"time"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

type Deleter interface {
	DeleteLinks(ctx context.Context, params *sclient.DeleteLinksParams, body sclient.DeleteLinksJSONRequestBody,
		reqEditors ...sclient.RequestEditorFn) (*http.Response, error)
}

type Untrack struct {
	Traits   *models.Traits
	Pipeline []*models.Stage
	Link     sclient.RemoveLinkRequest
	Client   Deleter
}

func NewUntrack(chatID int64, client Deleter) *Untrack {
	return &Untrack{
		Traits:   models.NewTraits(UntrackSpan, chatID, CommandUntrack),
		Pipeline: createUntrackStages(),
		Client:   client,
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

func (c *Untrack) Request() string {
	params := &sclient.DeleteLinksParams{TgChatId: c.Traits.ChatID}

	body := sclient.RemoveLinkRequest{Link: c.Link.Link}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.Client.DeleteLinks(ctx, params, body)
	if err != nil {
		return FailedUntrack
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return FailedUntrack
	}

	if resp.StatusCode == http.StatusConflict {
		return LinkNotYetTracked
	}

	return SuccessfulUntrack
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
