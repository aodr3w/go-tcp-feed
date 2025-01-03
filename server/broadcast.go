package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/aodr3w/go-chat/data"
)

type Broadcast struct {
	m       *sync.RWMutex
	data    [][]byte
	readIdx map[string]int
	dao     *data.Dao
}

func NewBroadCast(dao *data.Dao) *Broadcast {
	return &Broadcast{
		m:       &sync.RWMutex{},
		data:    make([][]byte, 0),
		readIdx: make(map[string]int),
		dao:     dao,
	}
}

func (bc *Broadcast) Write(d []byte) error {
	msg, err := data.PayloadFromBytes(d)
	if err != nil {
		return fmt.Errorf("error serializing message from bytes %w", err)
	}

	sender, err := bc.dao.GetUserByName(msg.Message.Name)

	if err != nil {
		return fmt.Errorf("error getting user associated with message %w", err)
	}

	return bc.dao.InsertUserMessage(sender.ID, msg.Message.Text)
}

func (bc *Broadcast) LoadMessages(offset int, size int, maxTime time.Time) ([]data.MessagePayload, error) {
	count, err := bc.dao.GetMessageCount()

	if err != nil {
		return nil, err
	}

	messages, err := bc.dao.GetMessages(size, offset, data.Oldest, maxTime)

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
