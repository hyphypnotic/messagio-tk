package services

import (
	"errors"
	"github.com/hyphypnotic/messagio-tk/internal/msgHandler/repositories"
)

type messageService struct {
	Repo repositories.MessageRepository
}

type MessageService interface {
	UpdateMessageStatusByID(id uint, status string) error
}

func NewMessageService(repo repositories.MessageRepository) MessageService {
	return &messageService{Repo: repo}
}

// validateMessageStatus проверяет, является ли статус сообщения корректным
func validateMessageStatus(status string) error {
	if status != "success" && status != "error" {
		return errors.New("message status must be 'success' or 'error'")
	}
	return nil
}

func (s *messageService) UpdateMessageStatusByID(id uint, status string) error {
	if err := validateMessageStatus(status); err != nil {
		return err
	}

	message, err := s.Repo.GetByID(id)
	if err != nil {
		return err
	}

	message.Status = status
	return s.Repo.Update(message)
}
