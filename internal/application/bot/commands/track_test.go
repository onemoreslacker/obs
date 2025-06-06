package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestTrackRequest(t *testing.T) {
	var chatID int64 = 1
	tags, filters := []string{}, []string{}
	tests := map[string]struct {
		link        string
		setupMocks  func(client *mocks.MockScrapperClient, cache *mocks.MockCache)
		expectedMsg string
		wantErr     bool
	}{
		"successful track (cache update)": {
			link: "https://github.com/example/repo",
			setupMocks: func(client *mocks.MockScrapperClient, cache *mocks.MockCache) {
				response := sclient.LinkResponse{
					Id:      1,
					Url:     "https://github.com/example/repo",
					Tags:    tags,
					Filters: filters,
				}

				buf := &bytes.Buffer{}
				require.NoError(t, json.NewEncoder(buf).Encode(response))

				client.On("PostLinks", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(buf),
				}, nil)

				cache.On("Add", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(nil)
			},
			expectedMsg: commands.SuccessfulTrack,
			wantErr:     false,
		},
		"failed track (bad request)": {
			link: "https://github.com/example/repo2",
			setupMocks: func(client *mocks.MockScrapperClient, cache *mocks.MockCache) {
				client.On("PostLinks", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil)
			},
			expectedMsg: commands.FailedTrack,
			wantErr:     true,
		},
		"link already tracked (conflict)": {
			link: "https://github.com/example/repo3",
			setupMocks: func(client *mocks.MockScrapperClient, cache *mocks.MockCache) {
				client.On("PostLinks", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(&http.Response{
					StatusCode: http.StatusConflict,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}, nil)
			},
			expectedMsg: commands.LinkAlreadyTracked,
			wantErr:     false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			client := mocks.NewMockScrapperClient(t)
			defer client.AssertExpectations(t)

			cache := mocks.NewMockCache(t)
			defer cache.AssertExpectations(t)

			test.setupMocks(client, cache)

			body := sclient.AddLinkRequest{
				Link:    test.link,
				Tags:    tags,
				Filters: filters,
			}

			cmd := &commands.Track{
				Traits: models.NewTraits(commands.TrackSpan, chatID, commands.CommandTrack),
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
