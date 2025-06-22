package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func TestListRequest(t *testing.T) {
	var chatID int64 = 1

	tests := map[string]struct {
		tags        []string
		filters     []string
		setupMocks  func(client *mocks.MockScrapperClient, cache *mocks.MockCache)
		expectedMsg string
		wantErr     bool
	}{
		"successful list (cache hit)": {
			tags:    []string{},
			filters: []string{},
			setupMocks: func(_ *mocks.MockScrapperClient, cache *mocks.MockCache) {
				response := sclient.ListLinksResponse{
					Links: []sclient.LinkResponse{
						{
							Id:      1,
							Url:     "https://github.com/example/repo",
							Tags:    []string{},
							Filters: []string{},
						},
					},
					Size: 1,
				}

				cache.On("Get", mock.Anything, chatID).
					Once().Return(response, nil)
			},
			expectedMsg: "1. https://github.com/example/repo\n",
			wantErr:     false,
		},
		"failed list (cache miss and client fail)": {
			tags:    []string{},
			filters: []string{},
			setupMocks: func(client *mocks.MockScrapperClient, cache *mocks.MockCache) {
				params := &sclient.GetLinksParams{TgChatId: chatID}

				client.On("GetLinks", mock.Anything, params).
					Once().Return(&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(nil),
				}, errors.New("failed to get links"))

				cache.On("Get", mock.Anything, chatID).
					Once().Return(sclient.ListLinksResponse{}, commands.ErrLinkNotExists)
			},
			expectedMsg: commands.FailedList,
			wantErr:     true,
		},
		"empty list (cache hit)": {
			tags:    []string{},
			filters: []string{},
			setupMocks: func(_ *mocks.MockScrapperClient, cache *mocks.MockCache) {
				response := sclient.ListLinksResponse{
					Links: []sclient.LinkResponse{},
					Size:  0,
				}

				cache.On("Get", mock.Anything, chatID).
					Once().Return(response, nil)
			},
			expectedMsg: commands.EmptyList,
			wantErr:     false,
		},
		"empty list (cache miss)": {
			tags:    []string{},
			filters: []string{},
			setupMocks: func(client *mocks.MockScrapperClient, cache *mocks.MockCache) {
				params := &sclient.GetLinksParams{TgChatId: chatID}
				response := sclient.ListLinksResponse{
					Links: []sclient.LinkResponse{},
					Size:  0,
				}

				buf := &bytes.Buffer{}
				require.NoError(t, json.NewEncoder(buf).Encode(response))
				client.On("GetLinks", mock.Anything, params).
					Once().Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(buf),
				}, nil)

				cache.On("Get", mock.Anything, chatID).
					Once().Return(sclient.ListLinksResponse{}, commands.ErrLinkNotExists)
			},
			expectedMsg: commands.EmptyList,
			wantErr:     false,
		},
		"successful list with sieved links (cache hit)": {
			tags:    []string{"go"},
			filters: []string{"example"},
			setupMocks: func(_ *mocks.MockScrapperClient, cache *mocks.MockCache) {
				response := sclient.ListLinksResponse{
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
				}

				cache.On("Get", mock.Anything, chatID).
					Once().Return(response, nil)
			},
			expectedMsg: "1. https://github.com/example/repo\n",
			wantErr:     false,
		},
		"successful list with sieved links (cache miss)": {
			tags:    []string{"go"},
			filters: []string{"example"},
			setupMocks: func(client *mocks.MockScrapperClient, cache *mocks.MockCache) {
				params := &sclient.GetLinksParams{TgChatId: chatID}
				response := sclient.ListLinksResponse{
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
				}

				buf := &bytes.Buffer{}
				require.NoError(t, json.NewEncoder(buf).Encode(response))
				client.On("GetLinks", mock.Anything, params).
					Once().Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(buf),
				}, nil)

				cache.On("Get", mock.Anything, chatID).
					Once().Return(response, commands.ErrLinkNotExists)

				cache.On("Add", mock.Anything, chatID, mock.Anything).
					Twice().Return(nil)
			},
			expectedMsg: "1. https://github.com/example/repo\n",
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

			cmd := &commands.List{
				Traits: models.NewTraits(commands.ListSpan, chatID, commands.CommandList),
				Client: client,
				Link: sclient.AddLinkRequest{
					Tags:    test.tags,
					Filters: test.filters,
				},
				Cache: cache,
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
