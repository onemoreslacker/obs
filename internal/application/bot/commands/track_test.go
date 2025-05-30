package commands_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/commands"
	mocks "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTrackRequest(t *testing.T) {
	var chatID int64 = 1
	tags, filters := []string{}, []string{}
	tests := map[string]struct {
		link        string
		statusCode  int
		expectedMsg string
	}{
		"successful track": {
			link:        "https://github.com/example/repo",
			statusCode:  http.StatusOK,
			expectedMsg: commands.SuccessfulTrack,
		},
		"failed track (bad request)": {
			link:        "https://github.com/example/repo2",
			statusCode:  http.StatusBadRequest,
			expectedMsg: commands.FailedTrack,
		},
		"link already tracked (conflict)": {
			link:        "https://github.com/example/repo3",
			statusCode:  http.StatusConflict,
			expectedMsg: commands.LinkAlreadyTracked,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := mocks.NewMockScrapperClient(t)
			defer client.AssertExpectations(t)

			params := &sclient.PostLinksParams{TgChatId: chatID}
			body := sclient.AddLinkRequest{
				Link:    test.link,
				Tags:    tags,
				Filters: filters,
			}

			client.On("PostLinks", mock.Anything, params, body).
				Return(&http.Response{
					StatusCode: test.statusCode,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil)

			cmd := &commands.Track{
				Traits: models.NewTraits(commands.TrackSpan, chatID, commands.CommandTrack),
				Client: client,
				Link:   body,
			}

			actualMsg := cmd.Request()

			require.Equal(t, test.expectedMsg, actualMsg)
		})
	}
}
