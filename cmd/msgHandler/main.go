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

// Message represents the structure of the messages being processed
type Message struct {
	ID     int    `json:"id"`
	Body   string `json:"body"`
	Status string `json:"status"`
}

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "messagio"
	dbPassword = "1234"
	dbName     = "messagio-tk"
)

func main() {
	// Database connection
	dbConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer closeDB(db)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Kafka consumer setup
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, "message_group", config)
	if err != nil {
		log.Fatalf("Failed to start Sarama consumer group: %v", err)
	}
	defer closeConsumerGroup(consumerGroup)

	consumer := &messageConsumer{db: db}

	// Create a context for consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consuming messages in a separate goroutine
	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{"message"}, consumer); err != nil {
				log.Fatalf("Error from consumer: %v", err)
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)
	<-sigterm

	log.Println("Terminating: received signal")
}

// closeDB closes the database connection
func closeDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Fatalf("Failed to close database connection: %v", err)
	}
}

// closeConsumerGroup closes the Kafka consumer group
func closeConsumerGroup(consumerGroup sarama.ConsumerGroup) {
	if err := consumerGroup.Close(); err != nil {
		log.Fatalf("Failed to close Kafka consumer group: %v", err)
	}
}

// messageConsumer implements the sarama.ConsumerGroupHandler interface
type messageConsumer struct {
	db *sql.DB
}

// Setup is called once at the beginning of a new session
func (consumer *messageConsumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called once at the end of a session
func (consumer *messageConsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from Kafka
func (consumer *messageConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var msg Message
		if err := json.Unmarshal(message.Value, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		log.Printf("Processing message: ID = %d, Body = %s, Status = %s, Timestamp = %v, Topic = %s",
			msg.ID, msg.Body, msg.Status, message.Timestamp, message.Topic)

		// Update message status in the database
		if _, err := consumer.db.Exec("UPDATE messages SET status = $1 WHERE id = $2", "success", msg.ID); err != nil {
			log.Printf("Failed to update message status: %v", err)
		}

		session.MarkMessage(message, "")
	}
	return nil
}
