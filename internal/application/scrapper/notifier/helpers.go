package notifier

import (
	"context"
	"fmt"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/pkg"
)

func (n *Notifier) SendUpdate(ctx context.Context, chatID int64, url, msg string) error {
	payload := botclient.PostUpdatesJSONRequestBody{
		Description: msg,
		Url:         url,
		TgChatId:    chatID,
	}

	resp, err := n.Sender.PostUpdates(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to post updates: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("post updates request failed, status: %w", err)
	}

	return nil
}

func (n *Notifier) Notify(ctx context.Context, chatID int64, url string) error {
	var (
		updates []models.Update
		err     error
	)

	service, err := pkg.ServiceFromURL(url)
	if err != nil {
		return err
	}

	switch service {
	case config.GitHub:
		updates, err = n.GitHub.RetrieveUpdates(ctx, url)
	case config.StackOverflow:
		updates, err = n.Stack.RetrieveUpdates(ctx, url)
	default:
		return fmt.Errorf("unsupported service: %s", service)
	}

	for _, update := range updates {
		if err := n.SendUpdate(ctx, chatID, url, update.String()); err != nil {
			return err
		}
	}

	return nil
}
