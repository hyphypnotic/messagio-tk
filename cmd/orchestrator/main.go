package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type Message struct {
	Id        uint64    `json:"id"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "messagio"
	password = "1234"
	dbname   = "messagio-tk"
)

func main() {
	e := echo.New()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	brokers := []string{"localhost:9092"}
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to start Sarama producer: %v", err)
	}
	defer producer.Close()

	e.POST("/", func(c echo.Context) error {
		message := new(Message)
		if err := c.Bind(message); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		// Сохранение сообщения в базу данных
		query := "INSERT INTO messages (body, status) VALUES ($1, $2) RETURNING id"
		err = db.QueryRow(query, message.Body, "pending").Scan(&message.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Отправка сообщения в Kafka
		jsonMsg, err := json.Marshal(message)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		kafkaMessage := &sarama.ProducerMessage{
			Topic: "message",
			Value: sarama.ByteEncoder(jsonMsg),
		}

		partition, offset, err := producer.SendMessage(kafkaMessage)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		log.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", "message", partition, offset)

		return c.JSON(http.StatusOK, message)
	})

	e.Logger.Fatal(e.Start(":5000"))
}
