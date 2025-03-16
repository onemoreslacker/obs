package commands

import (
	"net/url"
	"slices"
	"strings"
)

// TODO: carry out validation logic as to a separate package.

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
	// Не придумал, как по-человечески проверять валидность инпута..
	// Сначала хотел также кинуть в джсон, но как его сюда красиво прокинуть не понял.
	// Возможно всю логику валидации нужно в отдельный объект вынести.
	if !slices.Contains([]string{"yes", "no"},
		strings.ToLower(strings.TrimSpace(input))) {
		return ErrInvalidAck
	}

	return nil
}

func validateTags(input string) error {
	// NOTE: seems like this is always true.. omit validation?
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
