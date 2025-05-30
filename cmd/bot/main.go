package main

import (
	"context"

	binit "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/init"
	bs "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/service"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Decorate(
			binit.BotCommands,
		),
		fx.Provide(
			binit.Config,
			binit.TelegramAPI,
			binit.ScrapperClient,
			binit.Telebot,
			binit.BotServer,
			binit.BotService,
		),
		fx.Invoke(func(
			service *bs.BotService,
		) error {
			return service.Run(context.Background())
		}),
	)

	app.Run()
}
