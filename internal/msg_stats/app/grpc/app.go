package grpcapp

import (
	"context"
	"fmt"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc/reflection"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/msg_stats/services"
	"github.com/hyphypnotic/messagio-tk/internal/msg_stats/transport/grpc/message"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

type Server struct {
	logger     *zap.Logger
	config     *config.Config
	grpcServer *grpc.Server
}

func New(logger *zap.Logger, config *config.Config, msgStatsService services.MsgStats) *Server {
	grpcLogger := InterceptorLogger(logger)

	// Chain interceptors
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		recovery.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(grpcLogger),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		recovery.StreamServerInterceptor(),
		logging.StreamServerInterceptor(grpcLogger),
	}

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(unaryInterceptors...)),
		grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(streamInterceptors...)),
	}

	grpcServer := grpc.NewServer(serverOptions...)

	msgstatsgrpc.Register(grpcServer, msgStatsService)

	// for debugging
	if config.Env == "local" {
		reflection.Register(grpcServer)
	}

	return &Server{
		logger:     logger,
		config:     config,
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
