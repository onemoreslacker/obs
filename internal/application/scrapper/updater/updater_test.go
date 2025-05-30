package updater_test

import (
	"context"
	"testing"
	"time"

	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	mocks "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/updater"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProcessLink(t *testing.T) {
	storage := mocks.NewMockUpdaterStorage(t)
	defer storage.AssertExpectations(t)

	client := mocks.NewMockExternalClient(t)
	defer client.AssertExpectations(t)

	ctx := context.Background()

	links := []sapi.LinkResponse{
		{Id: 1, Url: "https://github.com/example/repo"},
	}

	storage.On("TouchLink", mock.Anything, int64(1)).Return(nil)
	storage.On("UpdateLinkActivity", mock.Anything, int64(1), true).Return(nil)

	client.On("RetrieveUpdates", mock.Anything, "https://github.com/example/repo").
		Return([]models.Update{{CreatedAt: time.Now().Format(time.RFC3339)}}, nil)

	upd := &updater.Updater{
		Storage: storage,
		GitHub:  client,
	}

	upd.ProcessLink(ctx, links)
}

func TestCheckActivity(t *testing.T) {
	client := mocks.NewMockExternalClient(t)
	defer client.AssertExpectations(t)
	ctx := context.Background()

	upd := &updater.Updater{
		GitHub: client,
	}

	client.On("RetrieveUpdates", mock.Anything, "https://github.com/example/repo").
		Return([]models.Update{{CreatedAt: time.Now().Format(time.RFC3339)}}, nil)

	active, err := upd.CheckActivity(ctx, "https://github.com/example/repo")
	require.NoError(t, err)
	require.True(t, active)
}
