package app

import (
	"database/sql"
	"fmt"
	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/services"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/transport/grpc"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Application struct {
	Config   *config.Config
	Logger   *zap.Logger
	Postgres *sql.DB
	Server   *grpc.Server
}

func New(cfg *config.Config, messageService services.MessageService) (*Application, error) {
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
	server, err := grpc.NewServer(cfg.GRPC.Address, messageService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
	}

	// Create and return the Application instance
	return &Application{
		Config:   cfg,
		Logger:   logger,
		Postgres: db,
		Server:   server,
	}, nil
}

func (app *Application) Run() error {
	// Start the gRPC server in a separate goroutine
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := app.Server.Start(); err != nil {
			app.Logger.Error("gRPC server stopped", zap.Error(err))
		}
	}()

	sig := <-shutdownCh
	app.Logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	app.Server.Stop()
	app.Logger.Info("gRPC server gracefully stopped")

	wg.Wait()

	return nil
}
