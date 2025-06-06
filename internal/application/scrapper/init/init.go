package scrapperinit

import (
	"flag"
	"fmt"
	"hash/adler32"
	"net"
	"net/http"
	"net/url"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	botclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/bot"
	sapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/notifier"
	ss "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/service"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/storage"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/updater"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/clients"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/producers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/chats"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/db"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/filters"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/links"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/subs"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/tags"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/txs"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/sender"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/servers/scrapper"
	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func Config() (*config.Config, error) {
	configFileName := flag.String("config", "", "path to config file")

	flag.Parse()

	cfg, err := config.New(*configFileName)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func DB(cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := db.New(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("pool was not created: %w", err)
	}

	return pool, nil
}

func ChatsRepository(cfg *config.Config, pool *pgxpool.Pool) chats.Repository {
	return chats.New(cfg.Database, pool)
}

func LinksRepository(cfg *config.Config, pool *pgxpool.Pool) links.Repository {
	return links.New(cfg.Database, pool)
}

func SubsRepository(cfg *config.Config, pool *pgxpool.Pool) subs.Repository {
	return subs.New(cfg.Database, pool)
}

func TagsRepository(cfg *config.Config, pool *pgxpool.Pool) tags.Repository {
	return tags.New(cfg.Database, pool)
}

func FiltersRepository(cfg *config.Config, pool *pgxpool.Pool) filters.Repository {
	return filters.New(cfg.Database, pool)
}

func Transactor(pool *pgxpool.Pool) *txs.TxBeginner {
	return txs.New(pool)
}

func Storage(
	chats chats.Repository,
	links links.Repository,
	subs subs.Repository,
	tags tags.Repository,
	filters filters.Repository,
	tx *txs.TxBeginner,
) *storage.Storage {
	return storage.New(chats, links, subs, tags, filters, tx)
}

func BotClient(cfg *config.Config) (botclient.ClientInterface, error) {
	server := url.URL{
		Scheme: config.Scheme,
		Host:   net.JoinHostPort(cfg.Serving.BotHost, cfg.Serving.BotPort),
	}

	client, err := botclient.NewClient(server.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create bot client: %w", err)
	}

	return client, nil
}

func StackClient() clients.Client {
	return clients.New(config.StackOverflow)
}

func GitHubClient() clients.Client {
	return clients.New(config.GitHub)
}

func Scheduler() (gocron.Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return scheduler, nil
}

func Serializer() *models.Serializer {
	return models.NewSerializer()
}

func Notifier(
	storage *storage.Storage,
	github clients.Client,
	stack clients.Client,
	sender *producers.UpdateSender,
	sch gocron.Scheduler,
	cfg *config.Config,
) *notifier.Notifier {
	return notifier.New(storage, github, stack, sender, sch, &cfg.Notifier)
}

func Updater(
	storage *storage.Storage,
	github clients.Client,
	stack clients.Client,
	sch gocron.Scheduler,
	cfg *config.Config,
) *updater.Updater {
	return updater.New(storage, github, stack, sch, &cfg.Updater)
}

func KafkaWriter(cfg *config.Config) *kafka.Writer {
	addresses := make([]string, 0, len(cfg.Brokers))
	for _, broker := range cfg.Brokers {
		addresses = append(addresses, net.JoinHostPort(broker.Host, broker.Port))
	}

	return &kafka.Writer{
		Addr:                   kafka.TCP(addresses...),
		Topic:                  cfg.Transport.Topic,
		Balancer:               &kafka.Hash{Hasher: adler32.New()},
		Transport:              kafka.DefaultTransport,
		AllowAutoTopicCreation: true,
	}
}

func AsyncSender(writer *kafka.Writer, serializer *models.Serializer) *producers.UpdateSender {
	return producers.NewUpdateSender(writer, serializer)
}

func UpdateSender(
	client botclient.ClientInterface,
	asyncSender *producers.UpdateSender,
	cfg *config.Config,
) (sender.UpdateSender, error) {
	syncSender := sender.NewSyncUpdateSender(client)

	snd, err := sender.New(syncSender, asyncSender, &cfg.Transport)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize update receiver: %w", err)
	}

	return snd, nil
}

func ScrapperServer(cfg *config.Config, storage *storage.Storage) *http.Server {
	api := sapi.New(storage)
	return scrapperserver.New(cfg.Serving, api)
}

func ScrapperService(
	upd *updater.Updater,
	nt *notifier.Notifier,
	sender *producers.UpdateSender,
	srv *http.Server,
) *ss.ScrapperService {
	return ss.New(upd, nt, sender, srv)
}
