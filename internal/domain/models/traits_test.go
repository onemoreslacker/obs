package models_test

import (
	"testing"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/stretchr/testify/require"
)

func TestTraitsTrack(t *testing.T) {
	tests := map[string]struct {
		initialStage  int
		input         string
		expectedLink  sclient.AddLinkRequest
		expectedStage int
	}{
		"Stage 0 sets link": {
			initialStage:  0,
			input:         "http://example.com",
			expectedLink:  sclient.AddLinkRequest{Link: "http://example.com"},
			expectedStage: 1,
		},
		"Stage 1 with 'no' skips tags": {
			initialStage:  1,
			input:         "no",
			expectedLink:  sclient.AddLinkRequest{Tags: []string{}},
			expectedStage: 3,
		},
		"Stage 1 with 'No' trimmed skips tags": {
			initialStage:  1,
			input:         "  No  ",
			expectedLink:  sclient.AddLinkRequest{Tags: []string{}},
			expectedStage: 3,
		},
		"Stage 1 with other input proceeds to tags": {
			initialStage:  1,
			input:         "yes",
			expectedLink:  sclient.AddLinkRequest{},
			expectedStage: 2,
		},
		"Stage 2 sets tags": {
			initialStage:  2,
			input:         "tag1 tag2",
			expectedLink:  sclient.AddLinkRequest{Tags: []string{"tag1", "tag2"}},
			expectedStage: 3,
		},
		"Stage 2 with empty input sets empty tags": {
			initialStage:  2,
			input:         "",
			expectedLink:  sclient.AddLinkRequest{Tags: []string{}},
			expectedStage: 3,
		},
		"Stage 3 with 'no' skips filters": {
			initialStage:  3,
			input:         "no",
			expectedLink:  sclient.AddLinkRequest{Filters: []string{}},
			expectedStage: 5,
		},
		"Stage 3 with other input proceeds to filters": {
			initialStage:  3,
			input:         "yes",
			expectedLink:  sclient.AddLinkRequest{},
			expectedStage: 4,
		},
		"Stage 4 sets filters": {
			initialStage:  4,
			input:         "filter1 filter2",
			expectedLink:  sclient.AddLinkRequest{Filters: []string{"filter1", "filter2"}},
			expectedStage: 5,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			traits := &models.Traits{Stage: test.initialStage}
			link := &sclient.AddLinkRequest{}
			traits.HandleTrack(test.input, link)

			require.Equal(t, *link, test.expectedLink)
			require.Equal(t, traits.Stage, test.expectedStage)
			require.Equal(t, false, traits.Malformed)
		})
	}
}

func TestTraitsUntrack(t *testing.T) {
	tests := map[string]struct {
		initialStage  int
		input         string
		expectedLink  sclient.RemoveLinkRequest
		expectedStage int
	}{
		"Stage 0 sets link": {
			initialStage:  0,
			input:         "http://example.com",
			expectedLink:  sclient.RemoveLinkRequest{Link: "http://example.com"},
			expectedStage: 1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			traits := &models.Traits{Stage: test.initialStage}
			link := &sclient.RemoveLinkRequest{}
			traits.HandleUntrack(test.input, link)

			require.Equal(t, *link, test.expectedLink)
			require.Equal(t, traits.Stage, test.expectedStage)
			require.Equal(t, false, traits.Malformed)
		})
	}
}

func TestTraitsList(t *testing.T) {
	tests := map[string]struct {
		initialStage  int
		input         string
		expectedLink  sclient.AddLinkRequest
		expectedStage int
	}{
		"Stage 1 with 'no' skips tags": {
			initialStage:  0,
			input:         "no",
			expectedLink:  sclient.AddLinkRequest{Tags: []string{}},
			expectedStage: 2,
		},
		"Stage 1 with 'No' trimmed skips tags": {
			initialStage:  0,
			input:         "  No  ",
			expectedLink:  sclient.AddLinkRequest{Tags: []string{}},
			expectedStage: 2,
		},
		"Stage 1 with other input proceeds to tags": {
			initialStage:  0,
			input:         "yes",
			expectedLink:  sclient.AddLinkRequest{},
			expectedStage: 1,
		},
		"Stage 2 sets tags": {
			initialStage:  1,
			input:         "tag1 tag2",
			expectedLink:  sclient.AddLinkRequest{Tags: []string{"tag1", "tag2"}},
			expectedStage: 2,
		},
		"Stage 2 with empty input sets empty tags": {
			initialStage:  1,
			input:         "",
			expectedLink:  sclient.AddLinkRequest{Tags: []string{}},
			expectedStage: 2,
		},
		"Stage 3 with 'no' skips filters": {
			initialStage:  2,
			input:         "no",
			expectedLink:  sclient.AddLinkRequest{Filters: []string{}},
			expectedStage: 4,
		},
		"Stage 3 with other input proceeds to filters": {
			initialStage:  2,
			input:         "yes",
			expectedLink:  sclient.AddLinkRequest{},
			expectedStage: 3,
		},
		"Stage 4 sets filters": {
			initialStage:  3,
			input:         "filter1 filter2",
			expectedLink:  sclient.AddLinkRequest{Filters: []string{"filter1", "filter2"}},
			expectedStage: 4,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			traits := &models.Traits{Stage: test.initialStage}
			link := &sclient.AddLinkRequest{}
			traits.HandleList(test.input, link)

			require.Equal(t, *link, test.expectedLink)
			require.Equal(t, traits.Stage, test.expectedStage)
			require.Equal(t, false, traits.Malformed)
		})
	}
}
