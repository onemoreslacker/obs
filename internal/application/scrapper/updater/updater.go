package updater

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"

	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/go-co-op/gocron/v2"
)

type ExternalClient interface {
	RetrieveUpdates(ctx context.Context, link string) ([]models.Update, error)
}

type Storage interface {
	GetLinks(ctx context.Context, batch uint64) ([]sapi.LinkResponse, error)
	TouchLink(ctx context.Context, linkID int64) error
	UpdateLinkActivity(ctx context.Context, linkID int64, status bool) error
}

type Updater struct {
	Storage Storage
	GitHub  ExternalClient
	Stack   ExternalClient
	Sch     gocron.Scheduler
	Cfg     updaterConfig
}

func New(
	storage Storage,
	github ExternalClient,
	stack ExternalClient,
	sch gocron.Scheduler,
	opts ...Option,
) *Updater {
	cfg := updaterConfig{
		batchSize:  1000,
		workersNum: runtime.GOMAXPROCS(0),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Updater{
		Storage: storage,
		GitHub:  github,
		Stack:   stack,
		Sch:     sch,
		Cfg:     cfg,
	}
}

func (upd *Updater) Run(ctx context.Context) {
	upd.ScrapeLinks(ctx)
}

func (upd *Updater) ScrapeLinks(ctx context.Context) {
	for {
		links, err := upd.Storage.GetLinks(ctx, upd.Cfg.batchSize)
		if err != nil {
			slog.Error(
				"updater: failed to get links",
				slog.String("msg", err.Error()),
			)
		}

		n := len(links)
		wg := sync.WaitGroup{}

		step := n / upd.Cfg.workersNum

		for off := 0; off < n; off += step {
			end := min(n, off+step)

			wg.Add(1)

			go func() {
				defer wg.Done()
				upd.ProcessLink(ctx, links[off:end])
			}()
		}

		wg.Wait()

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Minute):
		}
	}
}

func (upd *Updater) ProcessLink(ctx context.Context, batch []sapi.LinkResponse) {
	for _, link := range batch {
		updated, err := upd.CheckActivity(ctx, link.Url)
		if err != nil {
			slog.Error(
				"updater: failed to check link activity",
				slog.String("msg", err.Error()),
			)

			continue
		}

		if err := upd.Storage.TouchLink(ctx, link.Id); err != nil {
			slog.Error(
				"updater: failed to update link",
				slog.String("msg", err.Error()),
			)
			continue
		}

		if !updated {
			continue
		}

		if err := upd.Storage.UpdateLinkActivity(ctx, link.Id, updated); err != nil {
			slog.Error(
				"updater: failed to update link activity",
				slog.String("msg", err.Error()),
			)
		}
	}
}
