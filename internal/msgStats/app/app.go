package app

import (
	"database/sql"
	"fmt"
	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/app/grpc"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/repositories"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/services"
	"go.uber.org/zap"
)

type Application struct {
	Config     *config.Config
	Logger     *zap.Logger
	Postgres   *sql.DB
	GRPCServer *grpcapp.Server
}

func New(cfg *config.Config) (*Application, error) {
	// Initialize the logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize zap logger: %w", err)
	}

	// Set up the database connection
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

	// Initialize the gRPC server
	messageRepo := repositories.NewMessageRepo(db)
	msgStatsService := services.NewMsgStatsService(messageRepo)
	server := grpcapp.New(logger, cfg, msgStatsService)

	// Create and return the Application instance
	return &Application{
		Config:     cfg,
		Logger:     logger,
		Postgres:   db,
		GRPCServer: server,
	}, nil
}

func (app *Application) Run() error {
	err := app.GRPCServer.Run()
	return err
}
