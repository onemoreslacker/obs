package notifier_test

import (
	"context"
	"testing"
	"time"

	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/notifier"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNotify(t *testing.T) {
	client := mocks.NewMockExternalClient(t)
	defer client.AssertExpectations(t)

	sender := mocks.NewMockUpdateSender(t)
	defer sender.AssertExpectations(t)

	ctx := context.Background()

	client.On("RetrieveUpdates", mock.Anything, "https://github.com/example/repo").
		Once().Return([]models.Update{{CreatedAt: time.Now().Format(time.RFC3339)}}, nil)

	sender.On("Send", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).
		Once().Return(nil)

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

	sender := mocks.NewMockUpdateSender(t)
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

	sender.On("Send", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).
		Once().Return(nil)

	n := &notifier.Notifier{
		Storage: storage,
		GitHub:  client,
		Sender:  sender,
	}

	err := n.ProcessChat(ctx, 1)
	require.NoError(t, err)
}
