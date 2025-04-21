package notifier

import (
	"context"
	"fmt"
	"net/http"
	"time"

	botclient "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/bot"
)

func (n *Notifier) SendUpdates(ctx context.Context, chatID int64, url, msg string) error {
	payload := botclient.PostUpdatesJSONRequestBody{
		Description: &msg,
		Url:         &url,
		TgChatIds:   &[]int64{chatID},
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := n.sender.PostUpdates(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to post updates: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("post updates request failed, status: %w", err)
	}

	return nil
}
