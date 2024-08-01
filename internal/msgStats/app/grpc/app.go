package grpcapp

import (
	"context"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/services"
	msgstatsgrpc "github.com/hyphypnotic/messagio-tk/internal/msgStats/transport/grpc/message"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	logger     *zap.Logger
	config     *config.Config
	grpcServer *grpc.Server
}

func New(logger *zap.Logger, cfg *config.Config, msgStatsService services.MsgStats) *Server {
	grpcLogger := InterceptorLogger(logger)

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor(grpcLogger)),
		grpc.StreamInterceptor(logging.StreamServerInterceptor(grpcLogger)),
		grpc.UnaryInterceptor(recovery.UnaryServerInterceptor()),
		grpc.StreamInterceptor(recovery.StreamServerInterceptor()),
	}

	grpcServer := grpc.NewServer(serverOptions...)

	// Register msgStats service
	msgstatsgrpc.Register(grpcServer, msgStatsService)

	// Enable reflection for gRPC (optional, useful for debugging and tooling)
	reflection.Register(grpcServer)

	return &Server{
		logger:     logger,
		config:     cfg,
		grpcServer: grpcServer,
	}
}

func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		var zapLevel zapcore.Level
		switch lvl {
		case logging.LevelDebug:
			zapLevel = zap.DebugLevel
		case logging.LevelInfo:
			zapLevel = zap.InfoLevel
		case logging.LevelWarn:
			zapLevel = zap.WarnLevel
		case logging.LevelError:
			zapLevel = zap.ErrorLevel
		default:
			zapLevel = zap.InfoLevel
		}

		zapFields := make([]zap.Field, len(fields))
		for i, field := range fields {
			zapFields[i] = zap.Any(fmt.Sprintf("field_%d", i), field)
		}

		l.With(zapFields...).Check(zapLevel, msg).Write()
	})
}

func (server *Server) MustRun() {
	if err := server.Run(); err != nil {
		server.logger.Fatal("Server failed to start", zap.Error(err))
	}
}

func (server *Server) Run() error {
	const op = "grpcserver.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", server.config.GRPC.Port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	server.logger.Info("gRPC server started", zap.String("addr", l.Addr().String()))

	if err := server.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (server *Server) Stop() {
	const op = "grpcserver.Stop"

	server.logger.With(zap.String("op", op)).
		Info("stopping gRPC server", zap.Int("port", server.config.GRPC.Port))

	server.grpcServer.GracefulStop()
}
