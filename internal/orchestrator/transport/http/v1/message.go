package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/orchestrator/entity"
	"github.com/hyphypnotic/messagio-tk/internal/orchestrator/services"
	"github.com/hyphypnotic/messagio-tk/protos/gen/go/msgstats"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type messageRoutes struct {
	config   *config.Config
	service  services.MessageService
	grpcConn *grpc.ClientConn
	logger   *zap.Logger
}

type MessageController interface {
	Create(c echo.Context) error
	GetMsgStats(c echo.Context) error
}

func NewMessageRoutes(e *echo.Group, config *config.Config, service services.MessageService, grpcConn *grpc.ClientConn, logger *zap.Logger) {
	r := &messageRoutes{
		config:   config,
		service:  service,
		grpcConn: grpcConn,
		logger:   logger,
	}

	g := e.Group("/messages")
	{
		g.POST("/", r.Create)
		g.GET("/stats", r.GetMsgStats)
	}
}

func (r *messageRoutes) Create(c echo.Context) error {
	r.logger.Info("Received Create message request")
	message := new(entity.Message)
	if err := c.Bind(message); err != nil {
		r.logger.Error("Failed to bind request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload").SetInternal(err)
	}

	newMessage, err := r.service.Create(message)
	if err != nil {
		r.logger.Error("Failed to create message", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create message").SetInternal(err)
	}

	r.logger.Info("Message created successfully", zap.Uint("messageID", newMessage.ID))
	return c.JSON(http.StatusCreated, newMessage)
}

type MsgStatsRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func (r *messageRoutes) GetMsgStats(c echo.Context) error {
	r.logger.Info("Received GetMsgStats request")
	req := new(MsgStatsRequest)
	if err := c.Bind(req); err != nil {
		r.logger.Error("Failed to bind request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload").SetInternal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.config.GRPC.Timeout)*time.Second)
	defer cancel()

	msgStatsReq := &msgstats.MsgStatsRequest{
		StartTime: timestamppb.New(req.StartTime),
		EndTime:   timestamppb.New(req.EndTime),
	}
	msgStatsClient := msgstats.NewMsgStatsClient(r.grpcConn)
	resp, err := msgStatsClient.GetMsgStats(ctx, msgStatsReq)
	if err != nil {
		r.logger.Error("Failed to retrieve message stats", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve message stats").SetInternal(err)
	}

	r.logger.Info("Message stats retrieved successfully")
	return c.JSON(http.StatusOK, resp)
}
