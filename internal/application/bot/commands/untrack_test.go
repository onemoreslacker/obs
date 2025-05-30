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

func TestUntrackRequest(t *testing.T) {
	var chatID int64 = 1
	tests := map[string]struct {
		link        string
		statusCode  int
		expectedMsg string
	}{
		"successful untrack": {
			link:        "https://github.com/example/repo",
			statusCode:  http.StatusOK,
			expectedMsg: commands.SuccessfulUntrack,
		},
		"failed untrack (bad request)": {
			link:        "https://github.com/example/repo2",
			statusCode:  http.StatusBadRequest,
			expectedMsg: commands.FailedUntrack,
		},
		"link not yet tracked (conflict)": {
			link:        "https://github.com/example/repo3",
			statusCode:  http.StatusConflict,
			expectedMsg: commands.LinkNotYetTracked,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := mocks.NewMockScrapperClient(t)
			defer client.AssertExpectations(t)

			params := &sclient.DeleteLinksParams{TgChatId: chatID}
			body := sclient.RemoveLinkRequest{Link: test.link}

			client.On("DeleteLinks", mock.Anything, params, body).
				Return(&http.Response{
					StatusCode: test.statusCode,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil)

			cmd := &commands.Untrack{
				Traits: models.NewTraits(commands.UntrackSpan, 1, commands.CommandUntrack),
				Client: client,
				Link:   body,
			}
			actualMsg := cmd.Request()

			require.Equal(t, test.expectedMsg, actualMsg)
			client.AssertExpectations(t)
		})
	}
}
