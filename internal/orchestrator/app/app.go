package app

import (
	"database/sql"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/hyphypnotic/messagio-tk/internal/orchestrator/repositories"

	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/orchestrator/services"
	"github.com/hyphypnotic/messagio-tk/internal/orchestrator/transport/http/v1"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Application contains the application components
type Application struct {
	Config        *config.Config
	Logger        *zap.Logger
	Postgres      *sql.DB
	GRPCConn      *grpc.ClientConn
	Echo          *echo.Echo
	KafkaProducer sarama.SyncProducer
}

// New creates a new Application instance
func New(cfg *config.Config) (*Application, error) {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	// Set up database connection
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Error("Failed to open database", zap.Error(err))
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize gRPC connection (assuming it's created elsewhere)
	grpcAddress := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)

	grpcConn, err := grpc.NewClient(grpcAddress, grpc.WithInsecure())
	if err != nil {
		logger.Error("Failed to connect to gRPC server", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// Initialize Echo
	e := echo.New()
	kafkaProducer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, nil)
	if err != nil {
		return nil, err
	}

	return &Application{
		Config:        cfg,
		Logger:        logger,
		Postgres:      db,
		GRPCConn:      grpcConn,
		Echo:          e,
		KafkaProducer: kafkaProducer,
	}, nil
}

// Run starts the HTTP server
func (app *Application) Run() error {
	// Initialize routes
	msgRepo := repositories.NewMessageRepo(app.Postgres)
	v1.NewRouter(app.Echo, app.Config, app.Logger, app.GRPCConn, services.NewMessageService(msgRepo), app.KafkaProducer)

	// Start the server
	return app.Echo.Start(fmt.Sprintf(":%d", app.Config.HttpPort))
}
