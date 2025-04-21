package commands

import (
	"bytes"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
)

const (
	FiltersAck = "❔ Do you want to specify filters? (press /cancel to quit)"
	TagsAck    = "❔ Do you want to specify tags? (press /cancel to quit)"

	TagsRequest    = "✨ Please, enter link tags separated by space. (press /cancel to quit)"
	FiltersRequest = "✨ Please, enter link filters as filter:value. (press /cancel to quit)"

	LinkManual    = "💥 Invalid URL! Please enter a valid link (e.g. https://github.com/golang/go)"
	AcksManual    = "💥 Only yes/no are acceptable!"
	TagsManual    = "💥 Invalid tags! Use spaces to separate (e.g. 'work hobby')."
	FiltersManual = "💥 Invalid format! Use 'filter:value' (e.g. 'user:dummy')."
)

func constructListMessage(links []models.Link) string {
	var buf bytes.Buffer

	for i, link := range links {
		fmt.Fprintf(&buf, "%d. %s\n", i+1, *link.Url)
	}

	return buf.String()
}

func matchTags(got, desired []string) bool {
	if len(got) != len(desired) {
		return false
	}

	for _, tag := range desired {
		if !slices.Contains(got, tag) {
			return false
		}
	}

	return true
}

func matchFilters(got, desired []string) bool {
	if len(got) != len(desired) {
		return false
	}

	for _, filter := range desired {
		if !slices.Contains(got, filter) {
			return false
		}
	}

	return true
}

func ValidateLink(link string) error {
	_, err := url.Parse(link)
	if err != nil {
		return ErrInvalidLinkFormat
	}

	if !((strings.Contains(link, "stackoverflow") && strings.Contains(link, "questions")) ||
		strings.Contains(link, "github")) {
		return ErrInvalidLinkFormat
	}

	return nil
}

func validateAck(input string) error {
	if !slices.Contains([]string{"yes", "no"},
		strings.ToLower(strings.TrimSpace(input))) {
		return ErrInvalidAck
	}

	return nil
}

func validateTags(input string) error {
	if !(input == "" || len(strings.Fields(input)) > 0) {
		return ErrInvalidTagsFormat
	}

	return nil
}

func validateFilters(input string) error {
	filters := strings.Fields(input)

	for _, filter := range filters {
		pair := strings.Split(filter, ":")
		if len(pair) != 2 {
			return ErrInvalidFiltesFormat
		}
	}

	return nil
}
