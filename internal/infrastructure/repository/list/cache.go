package list

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/commands"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Rdb *redis.Client
}

func New(rdb *redis.Client) *Cache {
	return &Cache{
		Rdb: rdb,
	}
}

func (c *Cache) Add(ctx context.Context, chatID int64, link sclient.LinkResponse) error {
	key := strconv.FormatInt(chatID, 10)

	data, err := json.Marshal(link)
	if err != nil {
		return fmt.Errorf("cache: failed to serialize link: %w", err)
	}

	val, err := c.Rdb.HSet(ctx, key, link.Url, data).Result()
	if err != nil {
		return fmt.Errorf("cache: failed to push to hash: %w", err)
	}

	if val == 0 {
		return fmt.Errorf("cache: %w", commands.ErrLinkAlreadyExists)
	}

	return nil
}

func (c *Cache) Delete(ctx context.Context, chatID int64, req sclient.RemoveLinkRequest) error {
	key := strconv.FormatInt(chatID, 10)

	val, err := c.Rdb.HDel(ctx, key, req.Link).Result()
	if err != nil {
		return fmt.Errorf("cache: failed to delete from hash: %w", err)
	}

	if val == 0 {
		return fmt.Errorf("cache: %w", commands.ErrLinkNotExists)
	}

	return nil
}

func (c *Cache) Get(ctx context.Context, chatID int64) (sclient.ListLinksResponse, error) {
	key := strconv.FormatInt(chatID, 10)

	val, err := c.Rdb.HVals(ctx, key).Result()
	if err != nil {
		return sclient.ListLinksResponse{}, fmt.Errorf("cache: failed to get tracking links: %w", err)
	}

	links := make([]sclient.LinkResponse, 0, len(val))
	for _, link := range val {
		var decoded sclient.LinkResponse
		if err = json.Unmarshal([]byte(link), &decoded); err != nil {
			return sclient.ListLinksResponse{}, fmt.Errorf("cache: failed to serialize link: %w", err)
		}

		links = append(links, decoded)
	}

	return sclient.ListLinksResponse{
		Links: links,
		Size:  len(links),
	}, nil
}
