package entity

import "time"

type Message struct {
	ID        uint      `json:"id"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"create_at"`
}
