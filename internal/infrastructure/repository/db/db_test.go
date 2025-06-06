package db_test

import (
	"context"
	"log"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/chats"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/db"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/links"
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

func TestChatsRepository(t *testing.T) {
	tests := map[string]struct {
		access string
	}{
		"squirrel repository": {access: config.Orm},
		"sql repository":      {access: config.Sql},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			testDBConfig.Access = test.access
			repo := chats.New(testDBConfig, pool)

			require.NoError(t, clearChatsTable(ctx, pool))

			chatIDs := []int64{1, 2, 3, 4, 5, 6, 7, 8}

			t.Run("add chats", func(t *testing.T) {
				for _, chatID := range chatIDs {
					require.NoError(t, repo.Add(ctx, chatID))
				}
			})

			t.Run("chat exists", func(t *testing.T) {
				for _, chatID := range chatIDs {
					require.NoError(t, repo.ExistsID(ctx, chatID))
				}
			})

			t.Run("delete chats", func(t *testing.T) {
				retrievedIDs, err := repo.GetIDs(ctx)
				require.NoError(t, err)
				require.Equal(t, chatIDs, retrievedIDs)

				for _, chatID := range chatIDs {
					require.NoError(t, repo.Delete(ctx, chatID))
				}

				for _, chatID := range chatIDs {
					require.Error(t, repo.ExistsID(ctx, chatID))
				}

				retrievedIDs, err = repo.GetIDs(ctx)
				require.NoError(t, err)
				require.Equal(t, []int64{}, retrievedIDs)
			})
		})
	}
}

func TestLinksRepository(t *testing.T) {
	tests := map[string]struct {
		access string
	}{
		"squirrel repository": {access: config.Orm},
		"sql repository":      {access: config.Sql},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			testDBConfig.Access = test.access
			repo := links.New(testDBConfig, pool)

			require.NoError(t, clearLinksTable(ctx, pool))

			urls := []string{
				"https://github.com/example/repo",
				"https://github.com/example/repo2",
				"https://github.com/example/repo3",
				"https://github.com/example/repo4",
				"https://github.com/example/repo5",
				"https://github.com/example/repo6",
				"https://github.com/example/repo7",
				"https://github.com/example/repo8",
			}

			t.Run("add links", func(t *testing.T) {
				linkIDs := make([]int64, 0, len(urls))
				for _, url := range urls {
					linkID, err := repo.Add(ctx, url)
					require.NoError(t, err)
					linkIDs = append(linkIDs, linkID)
				}

				require.Equal(t, len(urls), len(linkIDs))
			})

			t.Run("get batch", func(t *testing.T) {
				batch, err := repo.GetBatch(ctx, uint64(len(urls)))
				require.NoError(t, err)
				require.Len(t, batch, len(urls))

				for _, link := range batch {
					require.True(t, slices.Contains(urls, link.Url))
				}
			})

			t.Run("delete links", func(t *testing.T) {
				batch, err := repo.GetBatch(ctx, uint64(len(urls)))
				require.NoError(t, err)
				require.Len(t, batch, len(urls))

				for _, link := range batch {
					require.NoError(t, repo.Delete(ctx, link.Id))
				}

				batch, err = repo.GetBatch(ctx, uint64(len(urls)))
				require.NoError(t, err)
				require.Len(t, batch, 0)
			})
		})
	}
}

func clearChatsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, "TRUNCATE TABLE chats CASCADE")
	return err
}

func clearLinksTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, "TRUNCATE TABLE links CASCADE")
	return err
}
