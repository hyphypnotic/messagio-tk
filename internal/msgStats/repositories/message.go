package repositories

import (
	"fmt"
	"time"

	"github.com/hyphypnotic/messagio-tk/internal/msgStats/app"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/entity"
)

type MessageRepo interface {
	CreateMessage(message *entity.Message) error
	GetMsgStats(startTime, endTime time.Time) (entity.MsgStats, error)
}

type messageRepo struct {
	App *app.Application
}

func NewMessageRepository(app *app.Application) MessageRepo {
	return &messageRepo{App: app}
}

func (r *messageRepo) CreateMessage(message *entity.Message) error {
	query := `INSERT INTO messages (body, status, created_at) VALUES ($1, $2, $3) RETURNING id`
	err := r.App.Postgres.QueryRow(query, message.Body, message.Status, time.Now()).Scan(&message.ID)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

func (r *messageRepo) GetMsgStats(startTime, endTime time.Time) (entity.MsgStats, error) {
	query := `
		SELECT 
			COUNT(*) AS total_count,
			COUNT(*) FILTER (WHERE status = 'success') AS success_count,
			COUNT(*) FILTER (WHERE status = 'error') AS error_count
		FROM messages
		WHERE created_at BETWEEN $1 AND $2
	`
	var totalCount, successCount, errorCount uint32
	err := r.App.Postgres.QueryRow(query, startTime, endTime).Scan(&totalCount, &successCount, &errorCount)
	if err != nil {
		return entity.MsgStats{}, fmt.Errorf("failed to get message statistics: %w", err)
	}

	msgStats := entity.MsgStats{
		Count: totalCount,
		Status: entity.Status{
			ErrorCount:   errorCount,
			SuccessCount: successCount,
		},
	}
	return msgStats, nil
}
