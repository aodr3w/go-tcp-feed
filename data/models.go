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

// represents a message sent by a client
type Message struct {
	ID        int
	userID    int
	Name      string
	Text      string
	CreatedAt time.Time
}

func (m *Message) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(m)
	if err != nil {
		return nil, fmt.Errorf("error encoding message: %w", err)
	}
	return buf.Bytes(), nil
}

// FromBytes converts a byte array to a Message struct using Gob.
func FromBytes(data []byte) (*Message, error) {
	var msg Message
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&msg)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %w", err)
	}
	return &msg, nil
}

func (m Message) String() string {
	return fmt.Sprintf(
		"Message{ID: %d, userID: %d, Name: %q, Text: %q, CreatedAt: %s}",
		m.ID, m.userID, m.Name, m.Text, m.CreatedAt.Format(time.RFC3339),
	)
}
