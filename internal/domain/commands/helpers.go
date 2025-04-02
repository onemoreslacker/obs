package commands

import (
	"bytes"
	"fmt"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
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

func constructListMessage(links []entities.Link) string {
	var buf bytes.Buffer

	for i, link := range links {
		fmt.Fprintf(&buf, "%d. %s\n", i+1, *link.Url)
	}

	return buf.String()
}
