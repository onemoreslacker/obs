package main

import (
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bootstrap"
	botservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/service"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Decorate(
			bootstrap.InitBotCommands,
		),
		fx.Provide(
			bootstrap.LoadConfig,
			bootstrap.InitTelegramAPI,
			bootstrap.InitScrapperClient,
			bootstrap.InitTelebot,
			bootstrap.InitBotServer,
			bootstrap.InitBotService,
		),
		fx.Invoke(func(
			service *botservice.BotService,
		) error {
			if err := service.Run(); err != nil {
				slog.Error(
					"Bot: service is down",
					slog.String("msg", err.Error()),
				)

				return err
			}

			return nil
		}),
	)

	app.Run()
}
