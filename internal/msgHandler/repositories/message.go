package repositories

import (
	"database/sql"
	"github.com/hyphypnotic/messagio-tk/internal/msgHandler/entity"
)

type MessageRepository interface {
	GetByID(id uint) (*entity.Message, error)
	Update(message *entity.Message) error
}

type messageRepository struct {
	DB *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{DB: db}
}

func (r *messageRepository) GetByID(id uint) (*entity.Message, error) {
	query := "SELECT id, body, created_at, status FROM messages WHERE id = $1"
	row := r.DB.QueryRow(query, id)

	var message entity.Message
	err := row.Scan(&message.ID, &message.Body, &message.CreatedAt, &message.Status)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) Update(message *entity.Message) error {
	query := "UPDATE messages SET body = $1, status = $2 WHERE id = $3"
	_, err := r.DB.Exec(query, message.Body, message.Status, message.ID)
	return err
}
