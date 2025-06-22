package main

import (
	"context"
	"os/signal"
	"syscall"

	sinit "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/init"
	ss "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			sinit.Config,
			sinit.DB,
			sinit.ChatsRepository,
			sinit.LinksRepository,
			sinit.SubsRepository,
			sinit.TagsRepository,
			sinit.FiltersRepository,
			sinit.Transactor,
			sinit.Storage,
			sinit.BotClient,
			sinit.Serializer,
			sinit.Limiter,
			fx.Annotate(
				sinit.StackClient,
				fx.ResultTags(`name:"stack"`),
			),
			fx.Annotate(
				sinit.GitHubClient,
				fx.ResultTags(`name:"github"`),
			),
			sinit.Scheduler,
			fx.Annotate(
				sinit.Notifier,
				fx.ParamTags(
					"",
					`name:"github"`,
					`name:"stack"`,
					"",
					"",
					"",
				),
			),
			fx.Annotate(
				sinit.Fetcher,
				fx.ParamTags(
					"",
					`name:"github"`,
					`name:"stack"`,
					"",
					"",
				),
			),
			sinit.Updater,
			sinit.KafkaUpdateWriter,
			sinit.UpdatePublisher,
			sinit.ScrapperServer,
			sinit.ScrapperService,
		),
		fx.Invoke(func(
			pool *pgxpool.Pool,
			updateWriter *kafka.Writer,
			service *ss.ScrapperService,
		) error {
			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			defer pool.Close()
			defer updateWriter.Close()

			return service.Run(ctx)
		}),
	)

	app.Run()
}
