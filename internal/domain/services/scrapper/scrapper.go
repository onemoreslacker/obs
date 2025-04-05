package scrapper

import (
	"log/slog"
	"net"
	"net/url"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/entities"
	botclient "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/external"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type Scrapper struct {
	botClient  botclient.ClientInterface
	repository linkRepository
	external   *external.Client
	cfg        *config.Config
	sched      gocron.Scheduler
}

type linkRepository interface {
	GetChatIDs() []int64
	GetLinks(int64) (links []entities.Link, err error)
}

func New(cfg *config.Config, repository linkRepository) (*Scrapper, error) {
	server := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(cfg.Serving.Host, cfg.Serving.BotPort),
	}

	client, err := botclient.NewClient(server.String())
	if err != nil {
		return nil, err
	}

	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &Scrapper{
		botClient:  client,
		repository: repository,
		external:   external.New(),
		cfg:        cfg,
		sched:      s,
	}, nil
}

func (s *Scrapper) Run() error {
	_, err := s.sched.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(10, 0, 0),
			),
		),
		gocron.NewTask(
			func() error {
				return s.scrapeUpdates()
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

	s.sched.Start()

	return nil
}
