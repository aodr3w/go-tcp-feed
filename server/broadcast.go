package server

import (
	"fmt"
	"log"
	"sync"

	"github.com/aodr3w/go-chat/db"
)

type Broadcast struct {
	m       *sync.RWMutex
	data    [][]byte
	readIdx map[string]int
	dao     *db.Dao
}

func NewBroadCast(dao *db.Dao) *Broadcast {
	return &Broadcast{
		m:       &sync.RWMutex{},
		data:    make([][]byte, 0),
		readIdx: make(map[string]int),
		dao:     dao,
	}
}

func (bc *Broadcast) Write(data []byte) error {
	msg, err := db.FromBytes(data)
	if err != nil {
		return fmt.Errorf("error serializing message from bytes %w", err)
	}
	log.Printf("received message: %v\n", msg)

	sender, err := bc.dao.GetUserByName(msg.Name)

	if err != nil {
		return fmt.Errorf("error getting user associated with message %w", err)
	}

	return bc.dao.InsertUserMessage(sender.ID, msg.Text)
}

func (bc *Broadcast) LoadMessages(offset int, size int) ([]db.Message, error) {
	messages, err := bc.dao.GetMessages(size, offset, db.Oldest)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (bc *Broadcast) Read(name string) []byte {
	bc.m.RLock()
	idx := bc.readIdx[name]
	defer bc.m.RUnlock()
	if idx >= 0 && idx < len(bc.data) {
		data := bc.data[idx]
		bc.readIdx[name] += 1
		if string(data)[:len(name)] != name {
			return data
		}
	}
	return []byte{}
}
