package server

import (
	"errors"
	"sync"
)

type Broadcast struct {
	m    *sync.RWMutex
	data [][]byte
}

func NewBroadCast() *Broadcast {
	return &Broadcast{
		m:    &sync.RWMutex{},
		data: make([][]byte, 0),
	}
}

func (bc *Broadcast) Write(data []byte) {
	bc.m.Lock()
	defer bc.m.Unlock()
	bc.data = append(bc.data, data)
}

func (bc *Broadcast) Read(idx int) ([]byte, error) {
	bc.m.RLock()
	defer bc.m.RUnlock()
	if idx >= 0 && idx < len(bc.data) {
		return bc.data[idx], nil
	}
	return nil, errors.New("idx out of bounds")
}
