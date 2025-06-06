package commands

import (
	"context"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

type List struct {
	Traits   *models.Traits
	Pipeline []*models.Stage
	Link     sclient.AddLinkRequest
	Client   Client
	Cache    Cache
}

func NewList(chatID int64, client Client, cache Cache) *List {
	return &List{
		Traits:   models.NewTraits(ListSpan, chatID, CommandList),
		Pipeline: createListStages(),
		Client:   client,
		Cache:    cache,
	}
}

func (c *List) Validate(input string) error {
	if err := c.Pipeline[c.Traits.Stage].Validate(input); err != nil {
		c.Traits.Malformed = true
		return err
	}

	c.Traits.HandleList(input, &c.Link)

	return nil
}

func (c *List) Stage() (string, bool) {
	keyboard := c.Traits.Stage == 0 || c.Traits.Stage == 2

	if !c.Traits.Malformed {
		return c.Pipeline[c.Traits.Stage].Prompt, keyboard
	}

	return c.Pipeline[c.Traits.Stage].Manual, keyboard
}

func (c *List) Done() bool {
	return c.Traits.Stage == c.Traits.Span
}

func (c *List) Request(ctx context.Context) (string, error) {
	list, err := c.GetLinksWithCache(ctx)
	if err != nil {
		return FailedList, err
	}

	sievedLinks := make([]sclient.LinkResponse, 0, list.Size)

	for _, Link := range list.Links {
		if MatchTags(Link.Tags, c.Link.Tags) &&
			MatchFilters(Link.Filters, c.Link.Filters) {
			sievedLinks = append(sievedLinks, Link)
		}
	}

	if len(sievedLinks) == 0 {
		return EmptyList, nil
	}

	return ConstructListMessage(sievedLinks), nil
}

func (c *List) Name() string {
	return c.Traits.Name
}

func createListStages() []*models.Stage {
	return []*models.Stage{
		models.NewStage(TagsAck, AcksManual, ValidateAck),
		models.NewStage(TagsRequest, TagsManual, ValidateTags),
		models.NewStage(FiltersAck, AcksManual, ValidateAck),
		models.NewStage(FiltersRequest, FiltersManual, ValidateFilters),
	}
}

const (
	CommandList = "list"
	ListSpan    = 4
	FailedList  = "💥 Failed to get the list of tracked Links."
	EmptyList   = "⚡️ Currently, there are no Links being tracked!"
)
