package updater

import (
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type Updater struct {
	sch        gocron.Scheduler
	updates    map[int64][]string
	repository LinksService
	external   ExternalClient
}

type ExternalClient interface {
	RetrieveStackOverflowUpdates(link string) ([]models.StackOverflowUpdate, error)
	RetrieveGitHubUpdates(link string) ([]models.GitHubUpdate, error)
}

type LinksService interface {
	GetChatIDs() ([]int64, error)
	GetLinks(int64) (links []models.Link, err error)
}

func New(repository LinksService) *Updater {
	return &Updater{
		repository: repository,
	}
}

func (upd *Updater) Run() error {
	_, err := upd.sch.NewJob(
		gocron.DurationJob(
			5*time.Minute,
		),
		gocron.NewTask(
			func() error {
				return upd.scrapeUpdates()
			},
		),
		gocron.WithEventListeners(
			gocron.AfterJobRunsWithError(
				func(jobID uuid.UUID, jobName string, err error) {
					slog.Error(
						"job error",
						slog.String("msg", err.Error()),
						slog.String("job_id", jobID.String()),
						slog.String("job_name", jobName),
						slog.String("service", "scrapper"),
					)
				},
			),
		),
	)

	if err != nil {
		return err
	}

	upd.sch.Start()

	return nil
}

func (upd *Updater) scrapeUpdates() error {
	chatIDs, err := upd.repository.GetChatIDs()
	if err != nil {
		return err
	}

	for _, chatID := range chatIDs {
		links, err := upd.repository.GetLinks(chatID)
		if err != nil {
			return err
		}

		slog.Info(
			"starting updates collection",
			slog.Int("total_chats", len(chatIDs)),
			slog.String("service", "scrapper"),
		)

		if err := upd.collectUpdates(links, chatID); err != nil {
			return err
		}
	}

	return nil
}

func (upd *Updater) collectUpdates(links []models.Link, chatID int64) error {
	for _, link := range links {
		l := *link.Url

		service, err := upd.identifyService(l)
		if err != nil {
			return err
		}

		switch service {
		case "github":
			updated, err := upd.checkForUpdatesGithub(l)
			if err != nil {
				return err
			}

			if !updated {
				continue
			}
		case "stackoverflow":
			updated, err := upd.checkForUpdatesStackOverflow(l)
			if err != nil {
				return err
			}

			if !updated {
				continue
			}
		}

		upd.updates[chatID] = append(upd.updates[chatID], l)

		slog.Info(
			"Found updates for link",
			"link", *link.Url,
			"service", service,
		)
	}

	return nil
}

func (upd *Updater) checkForUpdatesStackOverflow(link string) (bool, error) {
	updates, err := upd.external.RetrieveStackOverflowUpdates(link)
	if err != nil {
		return false, err
	}

	if len(updates) == 0 {
		return false, nil
	}

	createdAt := time.Unix(updates[0].CreatedAt, 0)

	return createdAt.After(getCutoff()), nil
}

func (upd *Updater) checkForUpdatesGithub(link string) (bool, error) {
	updates, err := upd.external.RetrieveGitHubUpdates(link)
	if err != nil {
		return false, err
	}

	if len(updates) == 0 {
		return false, nil
	}

	createdAt, err := time.Parse(time.RFC3339, updates[0].CreatedAt)
	if err != nil {
		return false, err
	}

	return createdAt.After(getCutoff()), nil
}

func (upd *Updater) identifyService(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	if strings.Contains(u.Host, "github") {
		return "github", nil
	}

	if strings.Contains(u.Host, "stackoverflow") {
		return "stackoverflow", nil
	}

	return "", ErrUnknownService
}

func getCutoff() time.Time {
	yesterday := time.Now().AddDate(0, 0, -1)
	cutoff := time.Date(
		yesterday.Year(),
		yesterday.Month(),
		yesterday.Day(),
		10,
		0,
		0,
		0,
		yesterday.Location(),
	)

	return cutoff
}
