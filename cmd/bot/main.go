package main

import (
	"context"
	"os/signal"
	"syscall"

	binit "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/init"
	bs "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/service"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
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
			binit.CircuitBreaker,
			binit.RoundTripper,
			binit.BotServer,
			binit.Deserializer,
			binit.Limiter,
			binit.KafkaUpdateReader,
			binit.KafkaDLQWriter,
			binit.ListRDB,
			binit.DLQPublisher,
			binit.UpdateSubscriber,
			binit.ListCache,
			binit.Processor,
			binit.BotService,
		),
		fx.Invoke(func(
			service *bs.BotService,
			dlqWriter *kafka.Writer,
			updateReader *kafka.Reader,
			listRDB *redis.Client,
		) error {
			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			defer dlqWriter.Close()
			defer updateReader.Close()
			defer listRDB.Close()

			return service.Run(ctx)
		}),
	)

	app.Run()
}
