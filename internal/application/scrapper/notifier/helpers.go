package notifier

import (
	"context"
	"fmt"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/svcident"
)

func (n *Notifier) Notify(ctx context.Context, chatID int64, link string) error {
	var (
		updates []models.Update
		err     error
	)

	service, err := svcident.FromLink(link)
	if err != nil {
		return err
	}

	switch service {
	case config.GitHub:
		updates, err = n.GitHub.RetrieveUpdates(ctx, link)
	case config.StackOverflow:
		updates, err = n.Stack.RetrieveUpdates(ctx, link)
	default:
		return fmt.Errorf("unsupported service: %s", service)
	}

	for _, update := range updates {
		if err = n.Sender.Send(ctx, chatID, link, update.String()); err != nil {
			return err
		}
	}

	return nil
}
