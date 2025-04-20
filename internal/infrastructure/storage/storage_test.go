package storage_test

import (
	"context"

	"log/slog"
	"math/rand/v2"
	"testing"
	"time"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/api/openapi/v1/scrapper_api"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDataInsertion(t *testing.T) {
	var (
		url     = "https://github.com/golang/go"
		tags    = []string{"tag"}
		filters = []string{"key:value"}
	)

	const (
		chatIDTagsFilters = iota
		chatIDTags
		chatIDFilters
		chatID
	)

	tests := map[string]struct {
		chatID int64
		link   models.Link
	}{
		"link with tags and filters insertion": {
			chatID: chatIDTagsFilters,
			link:   models.NewLink(rand.Int64(), url, tags, filters), //nolint:gosec // Temporary solution.
		},
		"link with tags insertion": {
			chatID: chatIDTags,
			link:   models.NewLink(rand.Int64(), url, tags, []string{}), //nolint:gosec // Temporary solution.
		},
		"link with filters insertion": {
			chatID: chatIDFilters,
			link:   models.NewLink(rand.Int64(), url, []string{}, filters), //nolint:gosec // Temporary solution.
		},
		"link without tags and filters": {
			chatID: chatID,
			link:   models.NewLink(rand.Int64(), url, []string{}, []string{}), //nolint:gosec // Temporary solution.
		},
	}

	repository := storage.NewLinksInMemoryService()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := repository.AddChat(test.chatID); err != nil {
				require.FailNow(t, err.Error())
			}

			if err := repository.AddLink(test.chatID, test.link); err != nil {
				require.FailNow(t, err.Error())
			}

			links, err := repository.GetChatLinks(test.chatID, true)
			if err != nil {
				require.FailNow(t, err.Error())
			}

			for _, link := range links {
				if *test.link.Id == *link.Id {
					require.Equal(t, test.link, link)
				}
			}
		})
	}
}

func TestHappyPath(t *testing.T) {
	var (
		url     = "https://github.com/golang/go"
		tags    = []string{"tag"}
		filters = []string{"key:value"}
	)

	const (
		chatIDTagsFilters = iota
		chatIDTags
		chatIDFilters
		chatID
	)

	tests := map[string]struct {
		chatID int64
		link   models.Link
	}{
		"link with tags and filters": {
			chatID: chatIDTagsFilters,
			link:   models.NewLink(rand.Int64(), url, tags, filters), //nolint:gosec // Temporary solution.
		},
		"link with tags": {
			chatID: chatIDTags,
			link:   models.NewLink(rand.Int64(), url, tags, []string{}), //nolint:gosec // Temporary solution.
		},
		"link with filters": {
			chatID: chatIDFilters,
			link:   models.NewLink(rand.Int64(), url, []string{}, filters), //nolint:gosec // Temporary solution.
		},
		"link without tags and filters": {
			chatID: chatID,
			link:   models.NewLink(rand.Int64(), url, []string{}, []string{}), //nolint:gosec // Temporary solution.
		},
	}

	repositories := storage.NewLinksInMemoryService()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := repositories.AddChat(test.chatID); err != nil {
				require.FailNow(t, err.Error())
			}

			if err := repositories.AddLink(test.chatID, test.link); err != nil {
				require.FailNow(t, err.Error())
			}

			if err := repositories.DeleteLink(test.chatID, *test.link.Url); err != nil {
				require.FailNow(t, err.Error())
			}
		})
	}
}

func TestDoubleInsertion(t *testing.T) {
	var (
		url     = "https://github.com/golang/go"
		tags    = []string{"tag"}
		filters = []string{"key:value"}
	)

	const (
		chatIDTagsFilters = iota
		chatIDTags
		chatIDFilters
		chatID
	)

	tests := map[string]struct {
		chatID int64
		link   models.Link
		want   error
	}{
		"link with tags and filters": {
			chatID: chatIDTagsFilters,
			link:   models.NewLink(rand.Int64(), url, tags, filters), //nolint:gosec // Temporary solution.
			want:   scrapperapi.ErrLinkAlreadyExists,
		},
		"link with tags": {
			chatID: chatIDTags,
			link:   models.NewLink(rand.Int64(), url, tags, []string{}), //nolint:gosec // Temporary solution.
			want:   scrapperapi.ErrLinkAlreadyExists,
		},
		"link with filters": {
			chatID: chatIDFilters,
			link:   models.NewLink(rand.Int64(), url, []string{}, filters), //nolint:gosec // Temporary solution.
			want:   scrapperapi.ErrLinkAlreadyExists,
		},
		"link without tags and filters": {
			chatID: chatID,
			link:   models.NewLink(rand.Int64(), url, []string{}, []string{}), //nolint:gosec // Temporary solution.
			want:   scrapperapi.ErrLinkAlreadyExists,
		},
	}

	repositories := storage.NewLinksInMemoryService()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := repositories.AddChat(test.chatID); err != nil {
				require.FailNow(t, err.Error())
			}

			if err := repositories.AddLink(test.chatID, test.link); err != nil {
				require.Equal(t, err, scrapperapi.ErrLinkAlreadyExists)
			}
		})
	}
}

func TestLinksService(t *testing.T) {
	ctx := context.Background()
	container, err := postgres.Run(
		ctx, "postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	require.NoError(t, err)

	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			slog.Error(
				"failed to terminate container",
				slog.String("msg", err.Error()),
			)
		}
	}()

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := config.Config{
		Database: config.Database{
			AccessType: "orm",
			Host:       host,
			Port:       port.Int(),
			Username:   "postgres",
			Password:   "postgres",
			Name:       "testdb",
		},
	}

	pool, err := storage.NewPoolWithMigrations(&cfg)
	require.NoError(t, err)

	implementations := []struct {
		name string
		cfg  config.Config
	}{
		{"ORM", config.Config{Database: config.Database{AccessType: "orm"}}},
		{"SQL", config.Config{Database: config.Database{AccessType: "sql"}}},
	}

	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			_, err := pool.Exec(ctx, "TRUNCATE TABLE chats, links, tracking_links CASCADE")
			require.NoError(t, err)

			service, err := storage.New(&impl.cfg, pool)
			require.NoError(t, err)

			testChatOperations(t, service)
			testLinkOperations(t, service)
		})
	}
}

func testChatOperations(t *testing.T, s storage.LinksService) {
	chatID := generateID()

	err := s.AddChat(chatID)
	require.NoError(t, err)

	err = s.AddChat(chatID)
	require.NotErrorIs(t, err, nil)

	ids, err := s.GetChatsIDs()
	require.NoError(t, err)
	require.Contains(t, ids, chatID)

	err = s.DeleteChat(chatID)
	require.NoError(t, err)

	_, err = s.GetChatLinks(chatID, true)
	require.NotErrorIs(t, err, nil)
}

func testLinkOperations(t *testing.T, s storage.LinksService) {
	chatID := generateID()
	require.NoError(t, s.AddChat(chatID))

	link := models.NewLink(generateID(), "https://github.com/onemoreslacker", []string{"tag"}, []string{"key:value"})
	require.NoError(t, s.AddLink(chatID, link))

	links, err := s.GetChatLinks(chatID, true)
	require.NoError(t, err)
	require.Len(t, links, 1)

	require.NoError(t, s.UpdateLinkActivity(*link.Id, true))

	require.NoError(t, s.TouchLink(*link.Id))

	require.NoError(t, s.DeleteLink(chatID, *link.Url))

	links, err = s.GetChatLinks(chatID, true)
	require.NoError(t, err)
	require.Len(t, links, 0)
}

func generateID() int64 {
	return int64(uuid.New().ID())
}
