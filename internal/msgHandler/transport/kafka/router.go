package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/hyphypnotic/messagio-tk/internal/config"
	"github.com/hyphypnotic/messagio-tk/internal/msgHandler/services"
	"go.uber.org/zap"
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
