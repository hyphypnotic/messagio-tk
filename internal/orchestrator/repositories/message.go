package repositories

import (
	"database/sql"
	"github.com/hyphypnotic/messagio-tk/internal/orchestrator/entity"
)

type messageRepo struct {
	DB *sql.DB
}

type MessageRepo interface {
	Create(message *entity.Message) (*entity.Message, error)
}

func NewMessageRepo(db *sql.DB) MessageRepo {
	return &messageRepo{DB: db}
}

func (repo *messageRepo) Create(message *entity.Message) (*entity.Message, error) {
	query := "INSERT INTO messages (body, status) VALUES ($1, $2) RETURNING id, created_at"
	err := repo.DB.QueryRow(query, message.Body, message.Status).Scan(&message.ID, &message.CreatedAt)
	if err != nil {
		return nil, err
	}
	return message, nil
}
