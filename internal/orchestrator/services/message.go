package services

import (
	"errors"
	"fmt"
	"github.com/hyphypnotic/messagio-tk/internal/orchestrator/entity"
	"github.com/hyphypnotic/messagio-tk/internal/orchestrator/repositories"
)

type messageService struct {
	repo repositories.MessageRepo
}
type MessageService interface {
	Create(message *entity.Message) (*entity.Message, error)
}

func NewMessageService(repo repositories.MessageRepo) MessageService {
	return &messageService{repo: repo}
}

func (s *messageService) Create(message *entity.Message) (*entity.Message, error) {
	if message == nil {
		return nil, errors.New("message cannot be nil")
	}
	if len(message.Body) == 0 {
		return nil, errors.New("message body cannot be empty")
	}

	createdMessage, err := s.repo.Create(message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return createdMessage, nil
}
