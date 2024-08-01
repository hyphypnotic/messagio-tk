package services

import (
	"fmt"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/app"
	"time"

	"github.com/hyphypnotic/messagio-tk/internal/msgStats/entity"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/repositories"
)

type MsgStats interface {
	CreateMessage(message *entity.Message) error
	GetMsgStats(startTime, endTime time.Time) (msgStats entity.MsgStats, err error)
}

type msgStats struct {
	app  *app.Application
	repo repositories.MessageRepo
}

func NewMessageService(repo repositories.MessageRepo) MsgStats {
	return &msgStats{repo: repo}
}

func (s *msgStats) CreateMessage(message *entity.Message) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}
	if message.Body == "" {
		return fmt.Errorf("message body cannot be empty")
	}
	if message.Status != "success" && message.Status != "error" {
		return fmt.Errorf("invalid message status: %s", message.Status)
	}

	return s.repo.CreateMessage(message)
}

// GetMsgStats retrieves message statistics for the specified time range.
func (s *msgStats) GetMsgStats(startTime, endTime time.Time) (msgStats entity.MsgStats, err error) {
	if startTime.After(endTime) {
		return entity.MsgStats{}, fmt.Errorf("start time (%v) must be before end time (%v)", startTime, endTime)
	}

	msgStats, err = s.repo.GetMsgStats(startTime, endTime)
	if err != nil {
		return entity.MsgStats{}, fmt.Errorf("failed to retrieve message statistics: %w", err)
	}

	return msgStats, nil
}
