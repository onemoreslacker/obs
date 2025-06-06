package main

import (
	"context"

	sinit "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/init"
	ss "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/service"
	"github.com/jackc/pgx/v5/pgxpool"
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
				),
			),
			fx.Annotate(
				sinit.Updater,
				fx.ParamTags(
					"",
					`name:"github"`,
					`name:"stack"`,
					"",
				),
			),
			sinit.KafkaWriter,
			sinit.AsyncSender,
			sinit.UpdateSender,
			sinit.ScrapperServer,
			sinit.ScrapperService,
		),
		fx.Invoke(func(
			pool *pgxpool.Pool,
			service *ss.ScrapperService,
		) error {
			defer pool.Close()
			return service.Run(context.Background())
		}),
	)

	app.Run()
}
