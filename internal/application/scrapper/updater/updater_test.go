package updater_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	bclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/updater"
	"github.com/es-debug/backend-academy-2024-go-template/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdater_Send_HTTPPrimaryKafkaFallback(t *testing.T) {
	ctx := context.Background()
	chatID := int64(12345)
	url := "https://example.com"
	description := "test description"

	tests := []struct {
		name            string
		transport       string
		httpError       error
		httpStatusCode  int
		kafkaError      error
		expectedError   bool
		expectHTTPCall  bool
		expectKafkaCall bool
	}{
		{
			name:            "http success - no fallback",
			transport:       config.HTTPTransport,
			httpError:       nil,
			httpStatusCode:  http.StatusOK,
			kafkaError:      nil,
			expectedError:   false,
			expectHTTPCall:  true,
			expectKafkaCall: false,
		},
		{
			name:            "http fails with error - fallback success",
			transport:       config.HTTPTransport,
			httpError:       errors.New("http connection failed"),
			httpStatusCode:  0,
			kafkaError:      nil,
			expectedError:   false,
			expectHTTPCall:  true,
			expectKafkaCall: true,
		},
		{
			name:            "http fails with bad status - fallback success",
			transport:       config.HTTPTransport,
			httpError:       nil,
			httpStatusCode:  http.StatusInternalServerError,
			kafkaError:      nil,
			expectedError:   false,
			expectHTTPCall:  true,
			expectKafkaCall: true,
		},
		{
			name:            "http fails - fallback fails",
			transport:       config.HTTPTransport,
			httpError:       errors.New("http connection failed"),
			httpStatusCode:  0,
			kafkaError:      errors.New("kafka connection failed"),
			expectedError:   true,
			expectHTTPCall:  true,
			expectKafkaCall: true,
		},
		{
			name:            "kafka success - no fallback",
			transport:       config.KafkaTransport,
			httpError:       nil,
			httpStatusCode:  http.StatusOK,
			kafkaError:      nil,
			expectedError:   false,
			expectHTTPCall:  false,
			expectKafkaCall: true,
		},
		{
			name:            "kafka fails - fallback success",
			transport:       config.KafkaTransport,
			httpError:       nil,
			httpStatusCode:  http.StatusOK,
			kafkaError:      errors.New("kafka connection failed"),
			expectedError:   false,
			expectHTTPCall:  true,
			expectKafkaCall: true,
		},
		{
			name:            "kafka fails - fallback fails",
			transport:       config.KafkaTransport,
			httpError:       errors.New("http connection failed"),
			httpStatusCode:  0,
			kafkaError:      errors.New("kafka connection failed"),
			expectedError:   true,
			expectHTTPCall:  true,
			expectKafkaCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpSenderMock := mocks.NewMockHTTPSender(t)
			kafkaSenderMock := mocks.NewMockKafkaSender(t)

			if tt.expectHTTPCall {
				expectedBody := bclient.PostUpdatesJSONRequestBody{
					TgChatId:    chatID,
					Url:         url,
					Description: description,
				}

				if tt.httpError != nil {
					httpSenderMock.On("PostUpdates", mock.Anything, expectedBody, mock.Anything).
						Return(&http.Response{}, tt.httpError)
				} else {
					mockResp := &http.Response{
						StatusCode: tt.httpStatusCode,
						Body:       http.NoBody,
					}
					httpSenderMock.On("PostUpdates", mock.Anything, expectedBody, mock.Anything).
						Return(mockResp, nil)
				}
			}

			if tt.expectKafkaCall {
				if tt.kafkaError != nil {
					kafkaSenderMock.On("Send", mock.Anything, chatID, url, description).
						Return(tt.kafkaError)
				} else {
					kafkaSenderMock.On("Send", mock.Anything, chatID, url, description).
						Return(nil)
				}
			}

			upd := updater.New(httpSenderMock, kafkaSenderMock, tt.transport)

			// Execute
			err := upd.Send(ctx, chatID, url, description)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			httpSenderMock.AssertExpectations(t)
			kafkaSenderMock.AssertExpectations(t)
		})
	}
}
