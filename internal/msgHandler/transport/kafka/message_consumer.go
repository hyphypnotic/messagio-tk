package kafka

import (
	"context"
	"encoding/json"
	"github.com/hyphypnotic/messagio-tk/internal/config"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/hyphypnotic/messagio-tk/internal/msgHandler/entity"
	"github.com/hyphypnotic/messagio-tk/internal/msgHandler/services"
)

type MessageConsumer struct {
	Cfg           *config.Config
	ConsumerGroup sarama.ConsumerGroup
	Service       services.MessageService
	Logger        *zap.Logger
}

const groupID = "message_group"

func NewMessageConsumer(service services.MessageService, logger *zap.Logger, appConfig *config.Config) *MessageConsumer {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(appConfig.Kafka.Brokers, groupID, kafkaConfig)
	if err != nil {
		logger.Fatal("Failed to start Sarama consumer group", zap.Error(err))
	}
	return &MessageConsumer{
		ConsumerGroup: consumerGroup,
		Service:       service,
		Logger:        logger,
		Cfg:           appConfig,
	}
}

func (c *MessageConsumer) StartConsuming(ctx context.Context) error {
	consumer := &messageConsumerHandler{
		Service: c.Service,
		Logger:  c.Logger,
	}
	return c.ConsumerGroup.Consume(ctx, c.Cfg.Kafka.Topics[groupID], consumer)
}

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
