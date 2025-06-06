package updater

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/config"
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
	Sem     chan struct{}
	Cfg     *config.Updater
}

func New(
	storage Storage,
	github ExternalClient,
	stack ExternalClient,
	sch gocron.Scheduler,
	cfg *config.Updater,
) *Updater {
	return &Updater{
		Storage: storage,
		GitHub:  github,
		Stack:   stack,
		Sch:     sch,
		Sem:     make(chan struct{}, cfg.NumWorkers),
		Cfg:     cfg,
	}
}

func (upd *Updater) Run(ctx context.Context) error {
	return upd.ScrapeLinks(ctx)
}

func (upd *Updater) ScrapeLinks(ctx context.Context) error {
	for {
		links, err := upd.Storage.GetLinks(ctx, upd.Cfg.BatchSize)
		if err != nil {
			slog.Warn(
				"updater: failed to get links",
				slog.String("msg", err.Error()),
			)
		}

		for _, link := range links {
			select {
			case <-ctx.Done():
				return context.Cause(ctx)
			case upd.Sem <- struct{}{}:
				go func() {
					defer func() { <-upd.Sem }()

					if err := upd.ProcessLink(ctx, link); err != nil {
						slog.Warn(
							"updater: failed to process link",
							slog.String("msg", err.Error()),
						)
					}
				}()
			}
		}

		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-time.After(5 * time.Minute):
		}
	}
}

func (upd *Updater) ProcessLink(ctx context.Context, link sapi.LinkResponse) error {
	updated, err := upd.CheckActivity(ctx, link.Url)
	if err != nil {
		return fmt.Errorf("failed to check activity: %w", err)
	}

	if !updated {
		return nil
	}

	if err := upd.Storage.UpdateLinkActivity(ctx, link.Id, updated); err != nil {
		return fmt.Errorf("failed to update link activity: %w", err)
	}

	if err := upd.Storage.TouchLink(ctx, link.Id); err != nil {
		return fmt.Errorf("failed to touch link: %w", err)
	}

	return nil
}
