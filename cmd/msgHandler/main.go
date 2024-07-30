package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq"
)

type Message struct {
	ID     int    `json:"id"`
	Body   string `json:"body"`
	Status string `json:"status"`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "messagio"
	password = "1234"
	dbname   = "messagio-tk"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, "message_group", config)
	if err != nil {
		log.Fatalf("Failed to start Sarama consumer group: %v", err)
	}
	defer func() {
		if err := consumerGroup.Close(); err != nil {
			log.Fatalf("Failed to close Sarama consumer group: %v", err)
		}
	}()

	consumer := Consumer{db: db}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{"message"}, &consumer); err != nil {
				log.Fatalf("Error from consumer: %v", err)
			}
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)
	<-sigterm

	log.Println("Terminating: via signal")
}

type Consumer struct {
	db *sql.DB
}

func (consumer *Consumer) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumer *Consumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var msg Message
		if err := json.Unmarshal(message.Value, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}
		log.Printf("Message claimed: id = %d, body = %s, status = %s, timestamp = %v, topic = %s", msg.ID, msg.Body, msg.Status, message.Timestamp, message.Topic)

		_, err := consumer.db.Exec("UPDATE messages SET status = $1 WHERE id = $2", "success", msg.ID)
		if err != nil {
			log.Printf("Failed to update message status: %v", err)
		}

		session.MarkMessage(message, "")
	}
	return nil
}
