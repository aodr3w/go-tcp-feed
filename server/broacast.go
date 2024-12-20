package server

import (
	"errors"
	"sync"
)

//NOTES
/*
since reading has a faster turn over time than writing, we should read first
and then attempt to write to the data slice in the handle connection function
*/
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

func (bc *Broadcast) Write(data []byte) int {
	bc.m.Lock()
	defer bc.m.Unlock()
	bc.data = append(bc.data, data)
	// Return the current length of the data slice, which represents the next index to read from.
	// This ensures the writer (publisher) can avoid reading the message it just wrote.
	// Consumers might encounter index errors if they try to read an index that is not yet written,
	// but this is expected behavior and should be handled by the caller.
	return len(bc.data)
}

func (bc *Broadcast) Read(idx int) ([]byte, error) {
	bc.m.RLock()
	defer bc.m.RUnlock()
	if idx >= 0 && idx < len(bc.data) {
		return bc.data[idx], nil
	}
	return nil, errors.New("idx out of bounds")
}
