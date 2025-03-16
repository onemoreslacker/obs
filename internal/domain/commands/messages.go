package commands

import (
	"bytes"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
)

func constructListMessage(links []entities.Link) string {
	buf := bytes.Buffer{}

	for _, link := range links {
		buf.Write([]byte(*link.Url + "\n"))
	}

	return buf.String()
}

func constructTags(input string) []string {
	return strings.Fields(input)
}

func constructFilters(input string) []string {
	return strings.Fields(input)
}
