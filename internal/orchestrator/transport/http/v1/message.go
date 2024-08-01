package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/hyphypnotic/messagio-tk/internal/orchestrator/app"
	"github.com/hyphypnotic/messagio-tk/protos/gen/go/msgstats"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Message represents the structure of a message.
type Message struct {
	ID        uint64    `json:"id"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// MsgStatsRequest represents the request payload for fetching message statistics.
type MsgStatsRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type Controller struct {
	App *app.Application
}

func NewController(app *app.Application) *Controller {
	return &Controller{App: app}
}

func (ctrl *Controller) PostMessage(c echo.Context) error {
	var msg Message
	if err := c.Bind(&msg); err != nil {
		ctrl.App.Logger.Warn("Invalid request payload", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	query := "INSERT INTO messages (body, status) VALUES ($1, $2) RETURNING id"
	if err := ctrl.App.Postgres.QueryRow(query, msg.Body, "pending").Scan(&msg.ID); err != nil {
		ctrl.App.Logger.Error("Failed to save message to database", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		ctrl.App.Logger.Error("Failed to marshal message", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	kafkaMessage := &sarama.ProducerMessage{
		Topic: ctrl.App.Config.Kafka.Topic,
		Value: sarama.ByteEncoder(msgJSON),
	}

	partition, offset, err := ctrl.App.KafkaProducer.SendMessage(kafkaMessage)
	if err != nil {
		ctrl.App.Logger.Error("Failed to send message to Kafka", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	ctrl.App.Logger.Info("Message sent to Kafka", zap.String("topic", ctrl.App.Config.Kafka.Topic), zap.Int32("partition", partition), zap.Int64("offset", offset))
	return c.JSON(http.StatusOK, msg)
}

func (ctrl *Controller) GetMessageStatistics(c echo.Context) error {
	var req MsgStatsRequest
	if err := c.Bind(&req); err != nil {
		ctrl.App.Logger.Warn("Invalid request payload", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	statsReq := &msgstats.MsgStatsRequest{
		StartTime: timestamppb.New(req.StartTime),
		EndTime:   timestamppb.New(req.EndTime),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := ctrl.App.GRPCClient.GetStats(ctx, statsReq)
	if err != nil {
		ctrl.App.Logger.Error("Failed to fetch message statistics", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	return c.JSON(http.StatusOK, resp)
}
