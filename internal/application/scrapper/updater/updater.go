package updater

import (
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
	repository LinksService
	external   ExternalClient
	sch        gocron.Scheduler
	workersNum int
}

type ExternalClient interface {
	RetrieveStackOverflowUpdates(link string) ([]models.StackOverflowUpdate, error)
	RetrieveGitHubUpdates(link string) ([]models.GitHubUpdate, error)
}

type LinksService interface {
	GetLinks(batchSize uint64) ([]models.Link, error)
	TouchLink(linkID int64) error
	UpdateLinkActivity(linkID int64, status bool) error
}

func New(repository LinksService, external ExternalClient, sch gocron.Scheduler) *Updater {
	return &Updater{
		repository: repository,
		external:   external,
		sch:        sch,
		workersNum: runtime.GOMAXPROCS(0),
	}
}

func (upd *Updater) Run() {
	upd.scrapeLinks()
}

func (upd *Updater) scrapeLinks() {
	var batchSize uint64 = 1000

	for {
		links, err := upd.repository.GetLinks(batchSize)
		if err != nil {
			slog.Error(
				"updater: GetLinks failed",
				slog.String("msg", err.Error()),
			)
		}

		n := len(links)
		wg := sync.WaitGroup{}

		step := n / upd.workersNum

		for off := 0; off < n; off += step {
			currentOff := off
			currentEnd := min(n, currentOff+step)

			wg.Add(1)

			go func(start, end int) {
				upd.processLink(links[start:end])
				wg.Done()
			}(currentOff, currentEnd)
		}

		wg.Wait()

		time.Sleep(5 * time.Minute)
	}
}

// processLink handles full lifecycle for a single link:
// 1. Checks for new activity using external APIs
// 2. Updates last checked timestamp
// 3. Updates activity status if changes detected
// Returns error only for fatal processing failures.
func (upd *Updater) processLink(batch []models.Link) {
	for _, link := range batch {
		if link.Url == nil {
			slog.Info("Updater: link's URL is missing")
			continue
		}

		updated, err := upd.checkActivity(*link.Url)
		if err != nil {
			slog.Error(
				"Updater: failed to check link activity",
				slog.String("msg", err.Error()),
			)

			continue
		}

		if link.Id == nil {
			slog.Info("Updater: link's id is missing")
			continue
		}

		if err := upd.repository.TouchLink(*link.Id); err != nil {
			slog.Error(
				"Updater: failed to update link",
				slog.String("msg", err.Error()),
			)

			continue
		}

		if updated {
			if err := upd.repository.UpdateLinkActivity(*link.Id, updated); err != nil {
				slog.Error(
					"Updater: failed to update link activity",
					slog.String("msg", err.Error()),
				)
			}
		}
	}
}
