package entity

import "time"

// Message represents the structure of a message.
type Message struct {
	ID        uint64    `json:"id"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
