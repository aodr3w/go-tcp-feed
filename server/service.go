package server

import (
	"fmt"
	"time"

	"github.com/aodr3w/go-tcp-feed/data"
)

type Service struct {
	*data.Dao
}

func NewService() *Service {
	d := data.NewDAO()
	return &Service{
		&d,
	}
}

func (s *Service) Write(d []byte) error {
	msg, err := data.PayloadFromBytes(d)
	if err != nil {
		return fmt.Errorf("error serializing message from bytes %w", err)
	}

	sender, err := s.GetUserByName(msg.Name)

	if err != nil {
		return fmt.Errorf("error getting user associated with message %w", err)
	}

	return s.InsertUserMessage(sender.ID, msg.Text, msg.CreatedAt)
}

func (s *Service) LoadMessages(offset int, size int, maxTime time.Time) ([]data.MessagePayload, error) {
	count, err := s.GetMessageCount()

	if err != nil {
		return nil, err
	}

	messages, err := s.GetMessages(size, offset, data.Oldest, maxTime)

	if err != nil {
		return nil, err
	}

	messagePayLoads := make([]data.MessagePayload, 0)
	for _, msg := range messages {
		messagePayLoads = append(messagePayLoads, data.MessagePayload{
			Message: msg,
			Count:   count,
		})
	}
	return messagePayLoads, nil
}
