package notifier_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	mocks "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/notifier"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPostUpdate(t *testing.T) {
	sender := mocks.NewMockSender(t)
	ctx := context.Background()

	sender.On("PostUpdates", mock.Anything, mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK}, nil)

	n := &notifier.Notifier{
		Sender: sender,
	}

	err := n.SendUpdate(ctx, 1, "https://github.com/example/repo", "update")
	require.NoError(t, err)

	sender.AssertExpectations(t)
}

func TestNotify(t *testing.T) {
	client := mocks.NewMockExternalClient(t)
	defer client.AssertExpectations(t)

	sender := mocks.NewMockSender(t)
	defer sender.AssertExpectations(t)

	ctx := context.Background()

	client.On("RetrieveUpdates", mock.Anything, "https://github.com/example/repo").
		Return([]models.Update{{CreatedAt: time.Now().Format(time.RFC3339)}}, nil)

	sender.On("PostUpdates", mock.Anything, mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK}, nil)

	n := &notifier.Notifier{
		GitHub: client,
		Sender: sender,
	}

	err := n.Notify(ctx, 1, "https://github.com/example/repo")
	require.NoError(t, err)
}

func TestProcessChat(t *testing.T) {
	storage := mocks.NewMockNotifierStorage(t)
	defer storage.AssertExpectations(t)

	client := mocks.NewMockExternalClient(t)
	defer client.AssertExpectations(t)

	sender := mocks.NewMockSender(t)
	defer sender.AssertExpectations(t)

	ctx := context.Background()

	links := []sapi.LinkResponse{
		{Id: 1, Url: "https://github.com/example/repo"},
	}

	storage.On("GetLinksWithChatActive", mock.Anything, mock.Anything).
		Return(links, nil)

	storage.On("UpdateLinkActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	client.On("RetrieveUpdates", mock.Anything, "https://github.com/example/repo").
		Return([]models.Update{{CreatedAt: time.Now().Format(time.RFC3339)}}, nil)

	sender.On("PostUpdates", mock.Anything, mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK}, nil)

	n := &notifier.Notifier{
		Storage: storage,
		GitHub:  client,
		Sender:  sender,
	}

	err := n.ProcessChat(ctx, 1)
	require.NoError(t, err)
}
