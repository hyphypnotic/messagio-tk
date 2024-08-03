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
	query := "INSERT INTO messages (body) VALUES ($1) RETURNING id, created_at, status"
	err := repo.DB.QueryRow(query, message.Body).Scan(&message.ID, &message.CreatedAt, &message.Status)
	if err != nil {
		return nil, err
	}
	return message, nil
}
