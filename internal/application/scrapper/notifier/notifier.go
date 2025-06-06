package notifier

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/go-co-op/gocron/v2"
)

type ExternalClient interface {
	RetrieveUpdates(ctx context.Context, link string) ([]models.Update, error)
}

type Storage interface {
	GetChatIDs(ctx context.Context) ([]int64, error)
	GetLinksWithChatActive(ctx context.Context, chatID int64) ([]sapi.LinkResponse, error)
	UpdateLinkActivity(ctx context.Context, linkID int64, status bool) error
}

type UpdateSender interface {
	Send(ctx context.Context, chatID int64, url, description string) error
}

type Notifier struct {
	Storage Storage
	GitHub  ExternalClient
	Stack   ExternalClient
	Sender  UpdateSender
	Sch     gocron.Scheduler
	Sem     chan struct{}
}

func New(
	storage Storage,
	github ExternalClient,
	stack ExternalClient,
	sender UpdateSender,
	sch gocron.Scheduler,
	cfg *config.Notifier,
) *Notifier {
	return &Notifier{
		Storage: storage,
		GitHub:  github,
		Stack:   stack,
		Sender:  sender,
		Sch:     sch,
		Sem:     make(chan struct{}, cfg.NumWorkers),
	}
}

func (n *Notifier) Run(ctx context.Context) error {
	_, err := n.Sch.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(10, 0, 0),
			),
		),
		gocron.NewTask(
			func() {
				n.PushUpdates(ctx)
			},
		),
	)

	if err != nil {
		return err
	}

	n.Sch.Start()

	return nil
}

func (n *Notifier) PushUpdates(ctx context.Context) {
	chatIDs, err := n.Storage.GetChatIDs(ctx)
	if err != nil {
		slog.Error(
			"notifier: failed to get chat ids",
			slog.String("msg", err.Error()),
		)
	}

	for _, chatID := range chatIDs {
		select {
		case <-ctx.Done():
			return
		case n.Sem <- struct{}{}:
			go func() {
				defer func() {
					<-n.Sem
				}()

				if err := n.ProcessChat(ctx, chatID); err != nil {
					slog.Error("notifier: processing failed",
						slog.Int64("chat_id", chatID),
						slog.String("error", err.Error()),
					)
				}
			}()
		}
	}
}

func (n *Notifier) ProcessChat(ctx context.Context, chatID int64) error {
	links, err := n.Storage.GetLinksWithChatActive(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat links: %w", err)
	}

	for _, link := range links {
		if err := n.Notify(ctx, chatID, link.Url); err != nil {
			return err
		}

		if err := n.Storage.UpdateLinkActivity(ctx, link.Id, false); err != nil {
			return err
		}
	}

	return nil
}
