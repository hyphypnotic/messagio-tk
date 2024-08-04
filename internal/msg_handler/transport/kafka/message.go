package kafka

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/hyphypnotic/messagio-tk/internal/msg_handler/entity"
	"github.com/hyphypnotic/messagio-tk/internal/msg_handler/services"
	"go.uber.org/zap"
)

type messageConsumerHandler struct {
	Service services.MessageService
	Logger  *zap.Logger
}

func (handler *messageConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (handler *messageConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (handler *messageConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var msg entity.Message
		if err := json.Unmarshal(message.Value, &msg); err != nil {
			handler.Logger.Error("Failed to unmarshal message", zap.Error(err))
			continue
		}

		handler.Logger.Info("Processing message",
			zap.Uint("ID", msg.ID),
			zap.String("Body", msg.Body),
			zap.String("Status", msg.Status),
			zap.Time("Timestamp", message.Timestamp),
			zap.String("Topic", message.Topic))

		// Update message status in the database
		if err := handler.Service.UpdateMessageStatusByID(msg.ID, "success"); err != nil {
			handler.Logger.Error("Failed to update message status", zap.Error(err))
		}

		session.MarkMessage(message, "")
	}
	return nil
}
