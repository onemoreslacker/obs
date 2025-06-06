package consumers_test

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	binit "github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/init"
	sinit "github.com/es-debug/backend-academy-2024-go-template/internal/application/scrapper/init"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/consumers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/producers"
	"github.com/es-debug/backend-academy-2024-go-template/internal/mocks"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

var (
	cfg         *config.Config
	kafkaReader *kafka.Reader
	kafkaWriter *kafka.Writer
	dlqWriter   *kafka.Writer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tckafka.Run(ctx,
		"confluentinc/cp-kafka:7.1.2",
		testcontainers.WithEnv(map[string]string{
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE": "true",
		}),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("failed to get container's host: %s", err)
	}

	port, err := container.MappedPort(ctx, "9093")
	if err != nil {
		log.Fatalf("failed to get container's port: %s", err)
	}

	brokerAddr := net.JoinHostPort(host, strconv.Itoa(port.Int()))

	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		log.Fatalf("failed to dial kafka: %s", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Fatalf("failed to get controller: %s", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Fatalf("failed to dial controller: %s", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             "link.updates",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             "link.updates.dlq",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Fatalf("failed to create topics: %s", err)
	}

	cfg = &config.Config{
		Brokers: []struct {
			Host string `yaml:"host"`
			Port string `yaml:"port"`
		}{
			{
				Host: host,
				Port: strconv.Itoa(port.Int()),
			},
		},
		Transport: config.Transport{
			Topic:           "link.updates",
			DLQTopic:        "link.updates.dlq",
			ConsumerGroupID: "link.updates.1",
		},
	}

	kafkaReader = binit.KafkaReader(cfg)
	dlqWriter = binit.KafkaWriter(cfg)
	kafkaWriter = sinit.KafkaWriter(cfg)

	os.Exit(m.Run())
}

func TestValidMessage(t *testing.T) {
	ctx := context.Background()

	dlqHandler := mocks.NewMockDLQHandler(t)
	defer dlqHandler.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)

	telegramClient := mocks.NewMockTgAPI(t)
	defer telegramClient.AssertExpectations(t)

	serializer := models.NewSerializer()
	deserializer := models.NewDeserializer()

	updateSender := producers.NewUpdateSender(kafkaWriter, serializer)
	updateReceiver := consumers.NewUpdateReceiver(kafkaReader, dlqHandler, telegramClient, deserializer)

	const messagesNumber = 5

	telegramClient.On("Send", mock.Anything).
		Times(messagesNumber).Return(tgbotapi.Message{}, nil)

	for range messagesNumber {
		chatID := int64(rand.Int())
		require.NoError(t, updateSender.Send(ctx, chatID,
			"https://github.com/example/repo", "something happened!"))
	}

	var processedMessages int
	for range messagesNumber {
		require.NoError(t, updateReceiver.ProcessMessage(ctx))
		processedMessages++
	}

	require.Equal(t, messagesNumber, processedMessages)
}

func TestDLQHandling(t *testing.T) {
	tests := map[string]struct {
		setupMocks         func(telegramClient *mocks.MockTgAPI, deserializer *mocks.MockDeserializer)
		chatID             int64
		deadLettersWritten int64
	}{
		"deserialization failed": {
			setupMocks: func(telegramClient *mocks.MockTgAPI, deserializer *mocks.MockDeserializer) {
				deserializer.On("Deserialize", mock.Anything, mock.Anything).
					Once().Return(errors.New("failed to deserialize update"))
			},
			chatID:             int64(rand.Int()),
			deadLettersWritten: 1,
		},
		"telegram client failed": {
			setupMocks: func(telegramClient *mocks.MockTgAPI, deserializer *mocks.MockDeserializer) {
				deserializer.On("Deserialize", mock.Anything, mock.Anything).
					Once().Return(nil)
				telegramClient.On("Send", mock.Anything).
					Once().Return(tgbotapi.Message{}, errors.New("failed to send a message"))
			},
			chatID:             int64(rand.Int()),
			deadLettersWritten: 1,
		},
		"successful processing": {
			setupMocks: func(telegramClient *mocks.MockTgAPI, deserializer *mocks.MockDeserializer) {
				deserializer.On("Deserialize", mock.Anything, mock.Anything).
					Once().Return(nil)
				telegramClient.On("Send", mock.Anything).
					Once().Return(tgbotapi.Message{}, nil)
			},
			chatID:             int64(rand.Int()),
			deadLettersWritten: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			telegramClient := mocks.NewMockTgAPI(t)
			defer telegramClient.AssertExpectations(t)

			deserializer := mocks.NewMockDeserializer(t)
			defer deserializer.AssertExpectations(t)

			test.setupMocks(telegramClient, deserializer)

			dlqHandler := producers.NewDLQHandler(dlqWriter)
			serializer := models.NewSerializer()
			updateSender := producers.NewUpdateSender(kafkaWriter, serializer)

			require.NoError(t, updateSender.Send(ctx, test.chatID,
				"https://github.com/example/repo", "something happened!"))

			updateReceiver := consumers.NewUpdateReceiver(kafkaReader, dlqHandler, telegramClient, deserializer)

			require.NoError(t, updateReceiver.ProcessMessage(ctx))
			require.Equal(t, test.deadLettersWritten, dlqWriter.Stats().Messages)
		})
	}
}
