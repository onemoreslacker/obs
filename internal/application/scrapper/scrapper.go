package scrapper

import (
	"log/slog"
	"net"
	"net/url"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/repositories"
	botclient "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients/external"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type Scrapper struct {
	botClient  botclient.ClientInterface
	repository repositories.LinksRepository
	external   *external.Client
	cfg        *config.Config
	sched      gocron.Scheduler
}

func New(cfg *config.Config, repository repositories.LinksRepository) (*Scrapper, error) {
	server := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(cfg.Host, cfg.BotPort),
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
		external:   external.New(cfg),
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
					slog.Info(
						"Job had an error",
						"job_id", jobID.String(),
						"job_name", jobName,
						"error", err.Error())
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
