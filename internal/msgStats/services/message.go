package services

import (
	"fmt"
	"time"

	"github.com/hyphypnotic/messagio-tk/internal/msgStats/entity"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/repositories"
)

type MsgStats interface {
	GetMsgStats(startTime, endTime time.Time) (msgStats entity.MsgStats, err error)
}

type msgStats struct {
	repo repositories.MessageRepo
}

func NewMsgStatsService(repo repositories.MessageRepo) MsgStats {
	return &msgStats{repo: repo}
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
