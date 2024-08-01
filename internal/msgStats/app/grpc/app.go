package grpcapp

import (
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/app"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/services"
	msgstatsgrpc "github.com/hyphypnotic/messagio-tk/internal/msgStats/transport/grpc/message"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	app        *app.Application
	gRPCServer *grpc.Server
}

func New(
	log *zap.Logger,
	msgStatsService services.MessageService,
	port int,
) *App {
	// Create a new gRPC server with Zap logger
	gRPCServer := grpc.NewServer()

	// Register msgStats service
	msgstatsgrpc.Register(gRPCServer, msgStatsService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}
