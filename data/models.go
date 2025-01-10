package data

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

// represents a connected client
type User struct {
	ID   int
	Name string
}

type MessagePayload struct {
	Message
	Count int
}

func NewMessage(text string, name string) Message {
	return Message{
		Name:      name,
		Text:      text,
		CreatedAt: time.Now(),
	}
}
func NewMessagePayload(m Message, count int) MessagePayload {
	return MessagePayload{
		Count:   count,
		Message: m,
	}
}

func (m *MessagePayload) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(m)
	if err != nil {
		return nil, fmt.Errorf("error encoding message: %w", err)
	}
	return buf.Bytes(), nil
}

// FromBytes converts a byte array to a Message struct using Gob.
func PayloadFromBytes(data []byte) (*MessagePayload, error) {
	var msg MessagePayload
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// represents a message sent by a client
type Message struct {
	ID        int
	userID    int
	Name      string
	Text      string
	CreatedAt time.Time
}

func (m MessagePayload) String() string {
	return fmt.Sprintf(
		"Message{ID: %d, userID: %d, Name: %q, Text: %q, CreatedAt: %s} count: %d",
		m.ID, m.userID, m.Name, m.Text, m.CreatedAt.Format(time.RFC3339), m.Count,
	)
}
