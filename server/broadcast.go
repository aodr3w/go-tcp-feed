package server

import (
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

func (bc *Broadcast) Write(data []byte) {
	bc.m.Lock()
	defer bc.m.Unlock()
	bc.data = append(bc.data, data)
}

func (bc *Broadcast) WriteV2(data []byte) error {
	//serialize data into message struct
	//must contain author information e.g userId
	//insert message into database
	return nil
}

func (bc *Broadcast) LoadMessages(offset int, size int) ([]db.Message, error) {
	/*called by client when chat is first open
	it loads message history
	*/
	messages, err := bc.dao.GetMessages(size, offset)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (bc *Broadcast) ReadMessages(userId int, offset int, size int) ([]db.Message, error) {
	/*retrieves all messages that were `sent` to the user*/
	messages, err := bc.dao.GetReceivedMessages(userId, size, offset)
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
