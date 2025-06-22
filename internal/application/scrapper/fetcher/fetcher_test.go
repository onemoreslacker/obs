package fetcher_test

import (
	"context"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/fetcher"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProcessLink(t *testing.T) {
	storage := mocks.NewMockUpdaterStorage(t)
	defer storage.AssertExpectations(t)

	client := mocks.NewMockExternalClient(t)
	defer client.AssertExpectations(t)

	ctx := context.Background()

	link := sapi.LinkResponse{Id: 1, Url: "https://github.com/example/repo"}

	storage.On("TouchLink", mock.Anything, int64(1)).Return(nil)
	storage.On("UpdateLinkActivity", mock.Anything, int64(1), true).Return(nil)

	client.On("RetrieveUpdates", mock.Anything, "https://github.com/example/repo").
		Return([]models.Update{{CreatedAt: time.Now().Format(time.RFC3339)}}, nil)

	upd := &fetcher.Fetcher{
		Storage: storage,
		GitHub:  client,
		Cfg: &config.Updater{
			BatchSize:  200,
			NumWorkers: 16,
		},
	}

	require.NoError(t, upd.ProcessLink(ctx, link))
}

func TestCheckActivity(t *testing.T) {
	client := mocks.NewMockExternalClient(t)
	defer client.AssertExpectations(t)
	ctx := context.Background()

	upd := &fetcher.Fetcher{
		GitHub: client,
	}

	client.On("RetrieveUpdates", mock.Anything, "https://github.com/example/repo").
		Return([]models.Update{{CreatedAt: time.Now().Format(time.RFC3339)}}, nil)

	active, err := upd.CheckActivity(ctx, "https://github.com/example/repo")
	require.NoError(t, err)
	require.True(t, active)
}
