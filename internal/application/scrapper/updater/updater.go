package updater

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/go-co-op/gocron/v2"
)

// Updater scrape links batches every n minutes and check for new activity.
// If new activity found, it updates last_activity_at timestamp.
type Updater struct {
	external   ExternalClient
	repository LinksService
	sch        gocron.Scheduler
	cfg        updaterConfig
}

type Option func(cfg *updaterConfig)

func WithCustomBatchSize(batchSize uint64) Option {
	return func(cfg *updaterConfig) {
		cfg.batchSize = batchSize
	}
}

func WithCustomWorkersNumber(workersNum int) Option {
	return func(cfg *updaterConfig) {
		cfg.workersNum = workersNum
	}
}

type updaterConfig struct {
	batchSize  uint64
	workersNum int
}

type ExternalClient interface {
	RetrieveStackOverflowUpdates(ctx context.Context, link string) ([]models.StackOverflowUpdate, error)
	RetrieveGitHubUpdates(ctx context.Context, link string) ([]models.GitHubUpdate, error)
}

type LinksService interface {
	GetLinks(ctx context.Context, batchSize uint64) ([]models.Link, error)
	TouchLink(ctx context.Context, linkID int64) error
	UpdateLinkActivity(ctx context.Context, linkID int64, status bool) error
}

// New instantiates a new Updater entity.
func New(repository LinksService, external ExternalClient, sch gocron.Scheduler, opts ...Option) *Updater {
	cfg := updaterConfig{
		batchSize:  1000,
		workersNum: runtime.GOMAXPROCS(0),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Updater{
		external:   external,
		repository: repository,
		sch:        sch,
		cfg:        cfg,
	}
}

func (upd *Updater) Run(ctx context.Context) {
	upd.scrapeLinks(ctx)
}

func (upd *Updater) scrapeLinks(ctx context.Context) {
	for {
		var (
			links []models.Link
			err   error
		)

		func() {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			links, err = upd.repository.GetLinks(ctx, upd.cfg.batchSize)
			if err != nil {
				slog.Error(
					"updater: GetLinks failed",
					slog.String("msg", err.Error()),
				)
			}
		}()

		n := len(links)
		wg := sync.WaitGroup{}

		step := n / upd.cfg.workersNum

		for off := 0; off < n; off += step {
			currentOff := off
			currentEnd := min(n, currentOff+step)

			wg.Add(1)

			go func(start, end int) {
				defer wg.Done()
				upd.processLink(ctx, links[start:end])
			}(currentOff, currentEnd)
		}

		wg.Wait()

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Minute):

		}
	}
}

// processLink handles full lifecycle for a single link:
// 1. Checks for new activity using external APIs
// 2. Updates last checked timestamp
// 3. Updates activity status if changes detected
// Returns error only for fatal processing failures.
func (upd *Updater) processLink(ctx context.Context, batch []models.Link) {
	for _, link := range batch {
		if link.Url == nil {
			slog.Info("updater: link's URL is missing")
			continue
		}

		updated, err := upd.checkActivity(ctx, *link.Url)
		if err != nil {
			slog.Error(
				"updater: failed to check link activity",
				slog.String("msg", err.Error()),
			)

			continue
		}

		if link.Id == nil {
			slog.Info("updater: link's id is missing")
			continue
		}

		if err := func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			return upd.repository.TouchLink(ctx, *link.Id)
		}(); err != nil {
			slog.Error(
				"updater: failed to update link",
				slog.String("msg", err.Error()),
			)
			continue
		}

		if updated {
			if err := func() error {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				return upd.repository.UpdateLinkActivity(ctx, *link.Id, updated)
			}(); err != nil {
				slog.Error(
					"updater: failed to update link activity",
					slog.String("msg", err.Error()),
				)
			}
		}
	}
}
