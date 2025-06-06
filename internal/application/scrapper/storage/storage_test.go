package storage_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/storage"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/chats"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/db"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/filters"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/links"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/subs"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/tags"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/txs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDBConfig config.Database
	pool         *pgxpool.Pool
)

func TestMain(m *testing.M) {
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
	defer func() {
		if err = testcontainers.TerminateContainer(container); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("failed to get container's host: %s", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("failed to get container's port: %s", err)
	}

	testDBConfig = config.Database{
		Host:     host,
		Port:     port.Int(),
		Username: "postgres",
		Password: "postgres",
		Name:     "testdb",
	}

	pool, err = db.New(testDBConfig)
	if err != nil {
		log.Fatalf("failed to initialize db: %s", err)
	}
	defer pool.Close()

	os.Exit(m.Run())
}

func TestAddAndGetLinks(t *testing.T) {
	tests := map[string]struct {
		access string
	}{
		"squirrel repository": {access: config.Orm},
		"sql repository":      {access: config.Sql},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			st := setupTestStorage(t, test.access)
			ctx := context.Background()

			testLinks := []struct {
				chatID int64
				link   sapi.AddLinkRequest
			}{
				{
					chatID: 1,
					link: sapi.AddLinkRequest{
						Link:    "https://github.com/example/repo",
						Tags:    []string{"go", "backend"},
						Filters: []string{"stars:>100", "license:mit"},
					},
				},
				{
					chatID: 1,
					link: sapi.AddLinkRequest{
						Link:    "https://github.com/example/repo2",
						Tags:    []string{"java", "backend"},
						Filters: []string{"stars:>1000", "license:mit"},
					},
				},
				{
					chatID: 2,
					link: sapi.AddLinkRequest{
						Link:    "https://github.com/example/repo3",
						Tags:    []string{"python", "ml"},
						Filters: []string{"stars:>500", "license:apache"},
					},
				},
			}

			for _, testLink := range testLinks {
				if err := st.ExistsChat(ctx, testLink.chatID); err != nil {
					require.NoError(t, st.AddChat(ctx, testLink.chatID))
				}

				_, err := st.AddLink(ctx, testLink.link, testLink.chatID)
				require.NoError(t, err)
			}

			t.Run("get links", func(t *testing.T) {
				links, err := st.GetLinksWithChat(ctx, 1)
				require.NoError(t, err)
				require.Len(t, links, 2)

				require.Equal(t, "https://github.com/example/repo", links[0].Url)
				require.ElementsMatch(t, []string{"go", "backend"}, links[0].Tags)
				require.ElementsMatch(t, []string{"stars:>100", "license:mit"}, links[0].Filters)

				require.Equal(t, "https://github.com/example/repo2", links[1].Url)
				require.ElementsMatch(t, []string{"java", "backend"}, links[1].Tags)
				require.ElementsMatch(t, []string{"stars:>1000", "license:mit"}, links[1].Filters)
			})

			t.Run("git active links", func(t *testing.T) {
				linkID, err := st.GetLinkID(ctx, "https://github.com/example/repo", 1)
				require.NoError(t, err)
				require.NoError(t, st.UpdateLinkActivity(ctx, linkID, true))

				activeLinks, err := st.GetLinksWithChatActive(ctx, 1)
				require.NoError(t, err)
				require.Len(t, activeLinks, 1)
				require.Equal(t, "https://github.com/example/repo", activeLinks[0].Url)
			})

			t.Run("get chat ids", func(t *testing.T) {
				chatIDs, err := st.GetChatIDs(ctx)
				require.NoError(t, err)
				require.ElementsMatch(t, []int64{1, 2}, chatIDs)
			})

			t.Run("repeated link", func(t *testing.T) {
				_, err := st.AddLink(ctx, testLinks[0].link, testLinks[0].chatID)
				require.Error(t, err)
			})
		})
	}
}

func TestLinkDeletion(t *testing.T) {
	tests := map[string]struct {
		access string
	}{
		"squirrel repository": {access: config.Orm},
		"sql repository":      {access: config.Sql},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			st := setupTestStorage(t, test.access)
			ctx := context.Background()

			chatID := int64(1)
			linkReq := sapi.AddLinkRequest{
				Link:    "https://github.com/example/to-delete",
				Tags:    []string{"test"},
				Filters: []string{"test:filter"},
			}

			if err := st.ExistsChat(ctx, chatID); err != nil {
				require.NoError(t, st.AddChat(ctx, chatID))
			}
			_, err := st.AddLink(ctx, linkReq, chatID)
			require.NoError(t, err)

			t.Run("successful deletion", func(t *testing.T) {
				deleteReq := sapi.RemoveLinkRequest{Link: linkReq.Link}
				require.NoError(t, st.DeleteLink(ctx, deleteReq, chatID))

				_, err := st.GetLinkID(ctx, linkReq.Link, chatID)
				require.Error(t, err)
				require.ErrorIs(t, err, sapi.ErrLinkNotExists)

				links, err := st.GetLinksWithChat(ctx, chatID)
				require.NoError(t, err)
				require.Empty(t, links)
			})

			t.Run("delete non-existent link", func(t *testing.T) {
				deleteReq := sapi.RemoveLinkRequest{Link: "https://github.com/non-existent"}
				err := st.DeleteLink(ctx, deleteReq, chatID)
				require.Error(t, err)
				require.ErrorIs(t, err, sapi.ErrLinkNotExists)
			})
		})
	}
}

func TestLinkActivity(t *testing.T) {
	tests := map[string]struct {
		access string
	}{
		"squirrel repository": {access: config.Orm},
		"sql repository":      {access: config.Sql},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			st := setupTestStorage(t, test.access)
			ctx := context.Background()

			chatID := int64(1)
			linkReq := sapi.AddLinkRequest{
				Link: "https://github.com/example/activity-test",
			}

			if err := st.ExistsChat(ctx, chatID); err != nil {
				require.NoError(t, st.AddChat(ctx, chatID))
			}
			linkID, err := st.AddLink(ctx, linkReq, chatID)
			require.NoError(t, err)

			t.Run("update activity", func(t *testing.T) {
				require.NoError(t, st.UpdateLinkActivity(ctx, linkID, true))
				activeLinks, err := st.GetLinksWithChatActive(ctx, chatID)
				require.NoError(t, err)
				require.Len(t, activeLinks, 1)

				require.NoError(t, st.UpdateLinkActivity(ctx, linkID, false))
				activeLinks, err = st.GetLinksWithChatActive(ctx, chatID)
				require.NoError(t, err)
				require.Empty(t, activeLinks)
			})
		})
	}
}

func setupTestStorage(t *testing.T, access string) *storage.Storage {
	ctx := context.Background()
	require.NoError(t, clearTables(ctx, pool))

	testDBConfig.Access = access

	return storage.New(
		chats.New(testDBConfig, pool),
		links.New(testDBConfig, pool),
		subs.New(testDBConfig, pool),
		tags.New(testDBConfig, pool),
		filters.New(testDBConfig, pool),
		txs.New(pool),
	)
}

func clearTables(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE 
			chats, 
			links, 
			subs, 
			tags, 
			filters 
		CASCADE
	`)
	return err
}
