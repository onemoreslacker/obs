package list_test

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/commands"
	botinit "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/init"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/list"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

var cache *list.Cache

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcredis.Run(ctx,
		"redis:8",
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
		return
	}

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("failed to get container's host: %s", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		log.Fatalf("failed to get container's port: %s", err)
	}

	cache = botinit.Cache(&config.Config{
		Cache: config.Cache{
			Host: host,
			Port: strconv.Itoa(port.Int()),
		},
	})

	os.Exit(m.Run())
}

func TestCache(t *testing.T) {
	ctx := context.Background()
	var chatID int64 = 1

	links := []sclient.LinkResponse{
		{
			Id:      1,
			Url:     "https://github.com/example/repo",
			Tags:    []string{"go", "backend"},
			Filters: []string{"stars:>100", "license:mit"},
		},
		{
			Id:      2,
			Url:     "https://github.com/example/repo2",
			Tags:    []string{"java", "backend"},
			Filters: []string{"stars:>1000", "license:mit"},
		},
		{
			Id:      3,
			Url:     "https://github.com/example/repo3",
			Tags:    []string{"python", "ml"},
			Filters: []string{"stars:>500", "license:apache"},
		},
		{
			Id:      4,
			Url:     "https://github.com/example/repo4",
			Tags:    []string{"scala", "backend"},
			Filters: []string{"stars:>1000", "license:apache"},
		},
		{
			Id:      5,
			Url:     "https://github.com/example/repo5",
			Tags:    []string{"c++", "embedded"},
			Filters: []string{"stars:>10000", "license:mit"},
		},
	}

	t.Run("add links", func(t *testing.T) {
		for _, link := range links {
			require.NoError(t, cache.Add(ctx, chatID, link))
		}
	})

	t.Run("get links", func(t *testing.T) {
		actualLinks, err := cache.Get(ctx, chatID)
		require.NoError(t, err)
		require.ElementsMatch(t, links, actualLinks.Links)
	})

	t.Run("delete links", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.NoError(t, cache.Delete(ctx, chatID, sclient.RemoveLinkRequest{Link: links[i].Url}))
		}
	})

	t.Run("get links after deleting", func(t *testing.T) {
		actualLinks, err := cache.Get(ctx, chatID)
		require.NoError(t, err)
		require.ElementsMatch(t, links[3:], actualLinks.Links)
	})

	t.Run("delete not existing link", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.ErrorIs(t, cache.Delete(ctx, chatID, sclient.RemoveLinkRequest{Link: links[i].Url}),
				commands.ErrLinkNotExists)
		}
	})

	t.Run("add already existing link", func(t *testing.T) {
		for i := 3; i < 5; i++ {
			require.ErrorIs(t, cache.Add(ctx, chatID, links[i]),
				commands.ErrLinkAlreadyExists)
		}
	})
}
