package models

import (
	"strings"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
)

type Traits struct {
	Stage     int
	Span      int
	ChatID    int64
	Malformed bool
	Name      string
}

func NewTraits(span int, chatID int64, name string) *Traits {
	return &Traits{
		Span:   span,
		ChatID: chatID,
		Name:   name,
	}
}

func (t *Traits) HandleTrack(input string, link *sclient.AddLinkRequest) {
	t.Malformed = false

	switch t.Stage {
	case 0:
		link.Link = input
	case 1:
		if strings.ToLower(strings.TrimSpace(input)) == "no" {
			t.Stage++
			link.Tags = []string{}
		}
	case 2:
		tags := strings.Fields(input)
		link.Tags = tags
	case 3:
		if strings.ToLower(strings.TrimSpace(input)) == "no" {
			t.Stage++
			link.Filters = []string{}
		}
	case 4:
		filters := strings.Fields(input)
		link.Filters = filters
	}

	t.Stage++
}

func (t *Traits) HandleUntrack(input string, link *sclient.RemoveLinkRequest) {
	t.Malformed = false

	link.Link = input

	t.Stage++
}

func (t *Traits) HandleList(input string, link *sclient.AddLinkRequest) {
	t.Malformed = false

	switch t.Stage {
	case 0:
		if strings.ToLower(strings.TrimSpace(input)) == "no" {
			t.Stage++
			link.Tags = []string{}
		}
	case 1:
		tags := strings.Fields(input)
		link.Tags = tags
	case 2:
		if strings.ToLower(strings.TrimSpace(input)) == "no" {
			t.Stage++
			link.Filters = []string{}
		}
	case 3:
		filters := strings.Fields(input)
		link.Filters = filters
	}

	t.Stage++
}
