package commands

import (
	"strings"
)

func constructTags(input string) []string {
	return strings.Fields(input)
}

func constructFilters(input string) []string {
	return strings.Fields(input)
}
