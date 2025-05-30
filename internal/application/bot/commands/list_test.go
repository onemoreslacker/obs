package commands_test

import (
	"bytes"
	"encoding/json"
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

func TestListRequest(t *testing.T) {
	var chatID int64 = 1

	tests := map[string]struct {
		tags        []string
		filters     []string
		response    sclient.ListLinksResponse
		statusCode  int
		expectedMsg string
	}{
		"successful track": {
			tags:    []string{},
			filters: []string{},
			response: sclient.ListLinksResponse{
				Links: []sclient.LinkResponse{
					{
						Id:      1,
						Url:     "https://github.com/example/repo",
						Tags:    []string{},
						Filters: []string{},
					},
				},
				Size: 1,
			},
			statusCode:  http.StatusOK,
			expectedMsg: "1. https://github.com/example/repo\n",
		},
		"failed list (bad request)": {
			tags:        []string{},
			filters:     []string{},
			statusCode:  http.StatusBadRequest,
			expectedMsg: commands.FailedList,
		},
		"empty list": {
			tags:    []string{},
			filters: []string{},
			response: sclient.ListLinksResponse{
				Links: []sclient.LinkResponse{},
				Size:  0,
			},
			statusCode:  http.StatusOK,
			expectedMsg: commands.EmptyList,
		},
		"successful list with sieved links": {
			tags:    []string{"go"},
			filters: []string{"example"},
			response: sclient.ListLinksResponse{
				Links: []sclient.LinkResponse{
					{
						Id:      1,
						Url:     "https://github.com/example/repo",
						Tags:    []string{"go"},
						Filters: []string{"example"},
					},
					{
						Id:      2,
						Url:     "https://gitlab.com/other/repo",
						Tags:    []string{"rust"},
						Filters: []string{"other"},
					},
				},
				Size: 2,
			},
			statusCode:  http.StatusOK,
			expectedMsg: "1. https://github.com/example/repo\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := mocks.NewMockScrapperClient(t)
			defer client.AssertExpectations(t)

			params := &sclient.GetLinksParams{TgChatId: chatID}
			body := &bytes.Buffer{}
			require.NoError(t, json.NewEncoder(body).Encode(test.response))

			client.On("GetLinks", mock.Anything, params).
				Return(&http.Response{
					StatusCode: test.statusCode,
					Body:       io.NopCloser(body),
				}, nil)

			cmd := &commands.List{
				Traits: models.NewTraits(commands.ListSpan, chatID, commands.CommandList),
				Client: client,
				Link: sclient.AddLinkRequest{
					Tags:    test.tags,
					Filters: test.filters,
				},
			}

			actualMsg := cmd.Request()

			require.Equal(t, test.expectedMsg, actualMsg)
		})
	}
}
