package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"sync"

	bclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/bot"
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

type Sender interface {
	PostUpdates(ctx context.Context, body bclient.PostUpdatesJSONRequestBody,
		reqEditors ...bclient.RequestEditorFn) (*http.Response, error)
}

type Notifier struct {
	Storage Storage
	GitHub  ExternalClient
	Stack   ExternalClient
	Sender  Sender
	Sch     gocron.Scheduler
	Guard   chan struct{}
}

func New(
	storage Storage,
	github ExternalClient,
	stack ExternalClient,
	sender Sender,
	sch gocron.Scheduler,
) *Notifier {
	return &Notifier{
		Storage: storage,
		GitHub:  github,
		Stack:   stack,
		Sender:  sender,
		Sch:     sch,
		Guard:   make(chan struct{}, runtime.GOMAXPROCS(0)),
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

	wg := sync.WaitGroup{}

	for _, chatID := range chatIDs {
		wg.Add(1)
		n.Guard <- struct{}{}

		go func(id int64) {
			defer func() {
				<-n.Guard
				wg.Done()
			}()

			if err := n.ProcessChat(ctx, id); err != nil {
				slog.Error("notifier: processing failed",
					slog.Int64("chat_id", id),
					slog.String("error", err.Error()))
			}
		}(chatID)
	}

	wg.Wait()
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
