package updater

import (
	"context"
	"fmt"
	"net/http"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	bclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/bot"
)

type HTTPSender interface {
	PostUpdates(ctx context.Context, body bclient.PostUpdatesJSONRequestBody,
		reqEditors ...bclient.RequestEditorFn) (*http.Response, error)
}

type KafkaSender interface {
	Send(ctx context.Context, chatID int64, url, description string) error
}

type Updater struct {
	httpSender  HTTPSender
	kafkaSender KafkaSender
	transport   string
}

func New(
	HTTPSender HTTPSender,
	KafkaSender KafkaSender,
	transport string,
) *Updater {
	return &Updater{
		httpSender:  HTTPSender,
		kafkaSender: KafkaSender,
		transport:   transport,
	}
}

func (u *Updater) Send(ctx context.Context, chatID int64, url, description string) error {
	var primaryErr, secondaryErr error

	switch u.transport {
	case config.HTTPTransport:
		if primaryErr = u.httpSend(ctx, chatID, url, description); primaryErr != nil {
			secondaryErr = u.kafkaSend(ctx, chatID, url, description)
		}
	case config.KafkaTransport:
		if primaryErr = u.kafkaSend(ctx, chatID, url, description); primaryErr != nil {
			secondaryErr = u.httpSend(ctx, chatID, url, description)
		}
	default:
		return ErrUnknownTransportMode
	}

	if primaryErr != nil && secondaryErr != nil {
		return fmt.Errorf("update sender error: %w", ErrSendUpdate)
	}

	return nil
}

func (u *Updater) httpSend(ctx context.Context, chatID int64, url, description string) error {
	resp, err := u.httpSender.PostUpdates(ctx, bclient.PostUpdatesJSONRequestBody{
		TgChatId:    chatID,
		Url:         url,
		Description: description,
	})
	if err != nil {
		return fmt.Errorf("failed to post updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrHTTPSendUpdate
	}

	return nil
}

func (u *Updater) kafkaSend(ctx context.Context, chatID int64, url, description string) error {
	return u.kafkaSender.Send(ctx, chatID, url, description)
}
