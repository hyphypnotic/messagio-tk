package v1

import (
	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/orchestrator/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewRouter(handler *echo.Echo, cfg *config.Config, l *zap.Logger, grpcConn *grpc.ClientConn, msgService services.MessageService) {
	handler.Use(middleware.Logger())
	handler.Use(middleware.Recover())

	v1 := handler.Group("/v1")
	{
		NewMessageRoutes(v1, cfg, msgService, grpcConn, l)
	}
}
