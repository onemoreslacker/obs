package botinit

import (
	"flag"
	"fmt"
	"hash/adler32"
	"net"
	"net/http"
	"net/url"

	"github.com/didip/tollbooth/v8/limiter"
	"github.com/es-debug/backend-academy-2024-go-template/config"
	sclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
	botapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/processor"
	botservice "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/service"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/telebot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/consumers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/producers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/list"
	botserver "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/servers/bot"
	cbt "github.com/es-debug/backend-academy-2024-go-template/pkg/cbtransport"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker/v2"
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

func CircuitBreaker(cfg *config.Config) *gobreaker.CircuitBreaker[*http.Response] {
	cbSettings := gobreaker.Settings{
		MaxRequests: cfg.CircuitBreakerPolicy.MaxRequests,
		Interval:    cfg.CircuitBreakerPolicy.Interval,
		Timeout:     cfg.CircuitBreakerPolicy.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.TotalFailures >= cfg.CircuitBreakerPolicy.FailureThreshold
		},
	}

	return gobreaker.NewCircuitBreaker[*http.Response](cbSettings)
}

func RoundTripper(cb *gobreaker.CircuitBreaker[*http.Response], cfg *config.Config) http.RoundTripper {
	return cbt.New(http.DefaultTransport, cb, &cbt.Config{
		MaxAttempts:    cfg.RetryPolicy.Attempts,
		RetryDelay:     cfg.RetryPolicy.Delay,
		RetryableCodes: cfg.RetryPolicy.StatusCodes,
	})
}

func ScrapperClient(transport http.RoundTripper, cfg *config.Config) (sclient.ClientInterface, error) {
	serverConn := url.URL{
		Scheme: config.Scheme,
		Host:   net.JoinHostPort(cfg.Serving.ScrapperHost, cfg.Serving.ScrapperPort),
	}

	client, err := sclient.NewClient(serverConn.String(), sclient.WithHTTPClient(&http.Client{
		Transport: transport,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create scrapper client: %w", err)
	}

	return client, nil
}

func BotServer(
	tgc *tgbotapi.BotAPI,
	cfg *config.Config,
	lmt *limiter.Limiter,
) *http.Server {
	api := botapi.New(tgc)
	return botserver.New(cfg, api, lmt)
}

func Telebot(client sclient.ClientInterface, tgc *tgbotapi.BotAPI, cache *list.Cache) *telebot.Bot {
	return telebot.New(client, tgc, cache)
}

func Deserializer() *models.Deserializer {
	return models.NewDeserializer()
}

func Limiter(cfg *config.Config) *limiter.Limiter {
	lmt := limiter.New(&limiter.ExpirableOptions{DefaultExpirationTTL: cfg.RateLimiter.TTL})
	lmt.SetMax(cfg.RateLimiter.MaxPerSecond)
	lmt.SetBurst(cfg.RateLimiter.Burst)
	lmt.SetMethods(cfg.RateLimiter.Methods)
	lmt.SetIPLookup(
		limiter.IPLookup{
			Name:           "RemoteAddr",
			IndexFromRight: 0,
		},
	)

	return lmt
}

func KafkaDLQWriter(cfg *config.Config) *kafka.Writer {
	addresses := make([]string, 0, len(cfg.Brokers))
	for _, broker := range cfg.Brokers {
		addresses = append(addresses, net.JoinHostPort(broker.Host, broker.Port))
	}

	return &kafka.Writer{
		Topic:                  cfg.Delivery.DLQTopic,
		Balancer:               &kafka.Hash{Hasher: adler32.New()},
		Transport:              kafka.DefaultTransport,
		AllowAutoTopicCreation: true,
	}
}

func KafkaUpdateReader(cfg *config.Config) *kafka.Reader {
	addresses := make([]string, 0, len(cfg.Brokers))
	for _, broker := range cfg.Brokers {
		addresses = append(addresses, net.JoinHostPort(broker.Host, broker.Port))
	}

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     addresses,
		Topic:       cfg.Delivery.Topic,
		StartOffset: kafka.FirstOffset,
		GroupID:     cfg.Delivery.ConsumerGroupID,
	})
}

func DLQPublisher(writer *kafka.Writer) *producers.DLQPublisher {
	return producers.NewDLQPublisher(writer)
}

func UpdateSubscriber(reader *kafka.Reader) *consumers.UpdateSubscriber {
	return consumers.NewUpdateSubscriber(reader)
}

func ListRDB(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Cache.Host, cfg.Cache.Port),
		Password: "",
		DB:       0,
	})
}

func ListCache(rdb *redis.Client) *list.Cache {
	return list.New(rdb)
}

func Processor(
	telegramSender *tgbotapi.BotAPI,
	deserializer *models.Deserializer,
	dlqPublisher *producers.DLQPublisher,
) *processor.Processor {
	return processor.New(telegramSender, deserializer, dlqPublisher)
}

func BotService(
	srv *http.Server,
	updateSubscriber *consumers.UpdateSubscriber,
	processor *processor.Processor,
	bot *telebot.Bot,
) *botservice.BotService {
	return botservice.New(srv, updateSubscriber, processor, bot)
}
