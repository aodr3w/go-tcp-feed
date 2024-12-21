package db

import "time"

// represents a connected client
type User struct {
	ID   int
	Name string
}

// represents a message sent by a client
type Message struct {
	ID        int
	Name      string
	Text      string
	CreatedAt time.Time
}
