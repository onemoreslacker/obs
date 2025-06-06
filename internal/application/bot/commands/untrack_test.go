package commands_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/commands"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUntrackRequest(t *testing.T) {
	tests := map[string]struct {
		link        string
		setupMocks  func(client *mocks.MockScrapperClient, cache *mocks.MockCache)
		expectedMsg string
		wantErr     bool
	}{
		"successful untrack": {
			link: "https://github.com/example/repo",
			setupMocks: func(client *mocks.MockScrapperClient, cache *mocks.MockCache) {
				client.On("DeleteLinks", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil)

				cache.On("Delete", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(nil)
			},
			expectedMsg: commands.SuccessfulUntrack,
			wantErr:     false,
		},
		"failed untrack (bad request)": {
			link: "https://github.com/example/repo2",
			setupMocks: func(client *mocks.MockScrapperClient, _ *mocks.MockCache) {
				client.On("DeleteLinks", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil)
			},
			expectedMsg: commands.FailedUntrack,
			wantErr:     true,
		},
		"link not yet tracked (conflict)": {
			link: "https://github.com/example/repo3",
			setupMocks: func(client *mocks.MockScrapperClient, _ *mocks.MockCache) {
				client.On("DeleteLinks", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(&http.Response{
					StatusCode: http.StatusConflict,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil)
			},
			expectedMsg: commands.LinkNotYetTracked,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			client := mocks.NewMockScrapperClient(t)
			defer client.AssertExpectations(t)

			cache := mocks.NewMockCache(t)
			defer client.AssertExpectations(t)

			test.setupMocks(client, cache)

			body := sclient.RemoveLinkRequest{Link: test.link}

			cmd := &commands.Untrack{
				Traits: models.NewTraits(commands.UntrackSpan, 1, commands.CommandUntrack),
				Client: client,
				Link:   body,
				Cache:  cache,
			}

			actualMsg, err := cmd.Request(ctx)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedMsg, actualMsg)
		})
	}
}
