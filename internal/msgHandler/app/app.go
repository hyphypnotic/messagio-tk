package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"

	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/msgHandler/transport/kafka"
)

type Application struct {
	Config        *config.Config
	Logger        *zap.Logger
	Postgres      *sql.DB
	KafkaProducer sarama.SyncProducer
	Consumer      *kafka.MessageConsumer
}

func NewApplication(cfg *config.Config) (*Application, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName)
	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Initialize Kafka producer
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Initialize Kafka consumer
	consumer := kafka.NewMessageConsumer(&Application{
		Config:        cfg,
		Logger:        logger,
		Postgres:      db,
		KafkaProducer: producer,
	})

	return &Application{
		Config:        cfg,
		Logger:        logger,
		Postgres:      db,
		KafkaProducer: producer,
		Consumer:      consumer,
	}, nil
}

func (app *Application) Close() {
	if err := app.Postgres.Close(); err != nil {
		app.Logger.Error("Failed to close PostgreSQL connection", zap.Error(err))
	}
	if err := app.KafkaProducer.Close(); err != nil {
		app.Logger.Error("Failed to close Kafka producer", zap.Error(err))
	}
	if err := app.Logger.Sync(); err != nil {
		log.Fatalf("Failed to sync logger: %v", err)
	}
}

func (app *Application) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := app.Consumer.StartConsuming(ctx); err != nil {
			app.Logger.Fatal("Failed to start consuming messages", zap.Error(err))
		}
	}()

	<-ctx.Done()
	app.Close()
}
