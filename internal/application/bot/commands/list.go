package commands

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	scrcl "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/scrapper"
)

type CommandList struct {
	traits *models.Traits

	pipeline []*models.Stage
	link     models.Link

	scrapperClient scrcl.ClientInterface
}

func NewCommandList(
	chatID int64,
	client scrcl.ClientInterface,
) *CommandList {
	return &CommandList{
		traits: models.NewTraits(
			listSpan,
			chatID,
			list,
		),
		pipeline:       createListStages(),
		scrapperClient: client,
	}
}

func (c *CommandList) Validate(input string) error {
	if err := c.pipeline[c.traits.Stage].Validate(input); err != nil {
		c.traits.Malformed = true
		return err
	}

	c.traits.HandleList(input, &c.link)

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

func (c *CommandList) Request() string {
	params := &scrcl.GetLinksParams{
		TgChatId: c.traits.ChatID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.scrapperClient.GetLinks(ctx, params)
	if err != nil {
		return failedList
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var respErr scrcl.ApiErrorResponse

		if err := json.NewDecoder(resp.Body).Decode(&respErr); err != nil {
			return failedList
		}

		if respErr.Description == nil {
			return failedList
		}

		return *respErr.Description
	}

	var list scrcl.ListLinksResponse

	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return failedList
	}

	sievedLinks := make([]models.Link, 0, *list.Size)

	for _, link := range *list.Links {
		if matchTags(*link.Tags, *c.link.Tags) &&
			matchFilters(*link.Filters, *c.link.Filters) {
			sievedLinks = append(sievedLinks, models.Link(link))
		}
	}

	if len(sievedLinks) == 0 {
		return emptyList
	}

	return constructListMessage(sievedLinks)
}

func (c *CommandList) Name() string {
	return c.traits.Name
}

func createListStages() []*models.Stage {
	return []*models.Stage{
		models.NewStage(
			TagsAck,
			AcksManual,
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
	list       = "list"
	listSpan   = 4
	failedList = "💥 Failed to get the list of tracked links."
	emptyList  = "⚡️ Currently, there are no links being tracked!"
)
