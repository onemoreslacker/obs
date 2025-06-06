package sender

import (
	"context"
	"fmt"
	"net/http"

	bclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/bot"
)

type Sender interface {
	PostUpdates(ctx context.Context, body bclient.PostUpdatesJSONRequestBody,
		reqEditors ...bclient.RequestEditorFn) (*http.Response, error)
}
type SyncAdapter struct {
	sender Sender
}

func NewSyncUpdateSender(sender Sender) *SyncAdapter {
	return &SyncAdapter{
		sender: sender,
	}
}

func (s *SyncAdapter) Send(ctx context.Context, chatID int64, url, description string) error {
	resp, err := s.sender.PostUpdates(ctx, bclient.PostUpdatesJSONRequestBody{
		TgChatId:    chatID,
		Url:         url,
		Description: description,
	})
	if err != nil {
		return fmt.Errorf("failed to post updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("post updates request failed, status: %w", err)
	}

	return nil
}

func (s *SyncAdapter) Stop() error { return nil }
