package main

import (
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bootstrap"
	scrapperservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			bootstrap.LoadConfig,
			bootstrap.InitPool,
			bootstrap.InitRepository,
			bootstrap.InitBotClient,
			bootstrap.InitExternalClient,
			bootstrap.InitTelegramAPI,
			bootstrap.InitScheduler,
			bootstrap.InitNotifier,
			bootstrap.InitUpdater,
			bootstrap.InitScrapperServer,
			bootstrap.InitScrapperService,
		),
		fx.Invoke(func(
			pool *pgxpool.Pool,
			service *scrapperservice.ScrapperService,
		) error {
			defer pool.Close()

			if err := service.Run(); err != nil {
				slog.Error(
					"scrapper service is down",
					slog.String("msg", err.Error()),
				)

				return err
			}

			return nil
		}),
	)

	app.Run()
}
