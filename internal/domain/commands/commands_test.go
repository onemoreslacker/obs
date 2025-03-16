package commands_test

import (
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/commands"
	"github.com/magiconair/properties/assert"
)

func TestLinkParsing(t *testing.T) {
	tests := map[string]struct {
		link string
		want error
	}{
		"unknown service link": {
			link: "https://reddit.com",
			want: commands.ErrInvalidLinkFormat,
		},
		"stackoverflow non-question link": {
			link: "https://stackoverflow.com/tags",
			want: commands.ErrInvalidLinkFormat,
		},
		"github valid link": {
			link: "https://github.com/golang/go",
			want: nil,
		},
		"stackoverflow valid link": {
			link: "https://stackoverflow.com/questions/1",
			want: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, commands.ValidateLink(test.link), test.want)
		})
	}
}
