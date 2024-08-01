package repositories

import (
	"fmt"
	"time"

	"github.com/hyphypnotic/messagio-tk/internal/msgStats/app"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/entity"
)

type MessageRepo interface {
	GetMsgStats(startTime, endTime time.Time) (entity.MsgStats, error)
}

type messageRepo struct {
	App *app.Application
}

func NewMessageRepo(app *app.Application) MessageRepo {
	return &messageRepo{App: app}
}

func (r *messageRepo) GetMsgStats(startTime, endTime time.Time) (entity.MsgStats, error) {
	var (
		query        string
		args         []interface{}
		totalCount   uint32
		successCount uint32
		errorCount   uint32
	)

	switch {
	case startTime.IsZero() && endTime.IsZero():
		query = `
			SELECT 
				COUNT(*) AS total_count,
				COUNT(*) FILTER (WHERE status = 'success') AS success_count,
				COUNT(*) FILTER (WHERE status = 'error') AS error_count
			FROM messages
		`
	case startTime.IsZero() && !endTime.IsZero():
		query = `
			SELECT 
				COUNT(*) AS total_count,
				COUNT(*) FILTER (WHERE status = 'success') AS success_count,
				COUNT(*) FILTER (WHERE status = 'error') AS error_count
			FROM messages
			WHERE created_at <= $1
		`
		args = append(args, endTime)
	case !startTime.IsZero() && endTime.IsZero():
		query = `
			SELECT 
				COUNT(*) AS total_count,
				COUNT(*) FILTER (WHERE status = 'success') AS success_count,
				COUNT(*) FILTER (WHERE status = 'error') AS error_count
			FROM messages
			WHERE created_at >= $1
		`
		args = append(args, startTime)
	default:
		query = `
			SELECT 
				COUNT(*) AS total_count,
				COUNT(*) FILTER (WHERE status = 'success') AS success_count,
				COUNT(*) FILTER (WHERE status = 'error') AS error_count
			FROM messages
			WHERE created_at BETWEEN $1 AND $2
		`
		args = append(args, startTime, endTime)
	}

	err := r.App.Postgres.QueryRow(query, args...).Scan(&totalCount, &successCount, &errorCount)
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
