package botinit

import (
	"flag"
	"fmt"
	"hash/adler32"
	"net"
	"net/http"
	"net/url"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	botapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/bot"
	botservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/service"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/telebot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/consumers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/producers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/receiver"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/list"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/servers/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func BotCommands(tgc *tgbotapi.BotAPI) error {
	commands := make([]tgbotapi.BotCommand, len(config.Descriptions))
	for i, data := range config.Descriptions {
		commands[i] = tgbotapi.BotCommand{
			Command:     data.Name,
			Description: data.Description,
		}
	}

	commandsConfig := tgbotapi.NewSetMyCommands(commands...)
	if _, err := tgc.Request(commandsConfig); err != nil {
		return fmt.Errorf("failed to load bot commands: %w", err)
	}

	return nil
}

func Config() (*config.Config, error) {
	configFileName := flag.String("config", "", "path to config file")

	flag.Parse()

	cfg, err := config.New(*configFileName)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func TelegramAPI(cfg *config.Config) (*tgbotapi.BotAPI, error) {
	tgc, err := tgbotapi.NewBotAPI(cfg.Secrets.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telegram api: %w", err)
	}

	return tgc, nil
}

func ScrapperClient(cfg *config.Config) (sclient.ClientInterface, error) {
	server := url.URL{
		Scheme: config.Scheme,
		Host:   net.JoinHostPort(cfg.Serving.ScrapperHost, cfg.Serving.ScrapperPort),
	}

	client, err := sclient.NewClient(server.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create scrapper client: %w", err)
	}

	return client, nil
}

func BotServer(tgc *tgbotapi.BotAPI, cfg *config.Config) *http.Server {
	api := botapi.New(tgc)
	return botserver.New(cfg.Serving, api)
}

func Telebot(client sclient.ClientInterface, tgc *tgbotapi.BotAPI, cache *list.Cache) *telebot.Bot {
	return telebot.New(client, tgc, cache)
}

func Deserializer() *models.Deserializer {
	return models.NewDeserializer()
}

func KafkaWriter(cfg *config.Config) *kafka.Writer {
	addresses := make([]string, 0, len(cfg.Brokers))
	for _, broker := range cfg.Brokers {
		addresses = append(addresses, net.JoinHostPort(broker.Host, broker.Port))
	}

	return &kafka.Writer{
		Addr:                   kafka.TCP(addresses...),
		Topic:                  cfg.Transport.DLQTopic,
		Balancer:               &kafka.Hash{Hasher: adler32.New()},
		Transport:              kafka.DefaultTransport,
		AllowAutoTopicCreation: true,
	}
}

func KafkaReader(cfg *config.Config) *kafka.Reader {
	addresses := make([]string, 0, len(cfg.Brokers))
	for _, broker := range cfg.Brokers {
		addresses = append(addresses, net.JoinHostPort(broker.Host, broker.Port))
	}

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     addresses,
		Topic:       cfg.Transport.Topic,
		StartOffset: kafka.FirstOffset,
		GroupID:     cfg.Transport.ConsumerGroupID,
	})
}

func DLQHandler(writer *kafka.Writer) *producers.DLQHandler {
	return producers.NewDLQHandler(writer)
}

func AsyncReceiver(
	reader *kafka.Reader,
	dlqHandler *producers.DLQHandler,
	tc *tgbotapi.BotAPI,
	deserializer *models.Deserializer,
) *consumers.UpdateReceiver {
	return consumers.NewUpdateReceiver(reader, dlqHandler, tc, deserializer)
}

func UpdateReceiver(
	srv *http.Server,
	asyncReceiver *consumers.UpdateReceiver,
	cfg *config.Config,
) (receiver.UpdateReceiver, error) {
	syncReceiver := receiver.NewSyncUpdateReceiver(srv)

	rcv, err := receiver.New(syncReceiver, asyncReceiver, &cfg.Transport)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize update receiver: %w", err)
	}

	return rcv, nil
}

func Cache(cfg *config.Config) *list.Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Cache.Host, cfg.Cache.Port),
		Password: "",
		DB:       0,
	})

	return list.New(rdb)
}

func BotService(
	rcv receiver.UpdateReceiver,
	bot *telebot.Bot,
) *botservice.BotService {
	return botservice.New(rcv, bot)
}
