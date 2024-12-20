package server

import (
	"sync"
)

//NOTES
/*
since reading has a faster turn over time than writing, we should read first
and then attempt to write to the data slice in the handle connection function
*/
type Broadcast struct {
	m       *sync.RWMutex
	data    [][]byte
	readIdx map[string]int
}

func NewBroadCast() *Broadcast {
	return &Broadcast{
		m:       &sync.RWMutex{},
		data:    make([][]byte, 0),
		readIdx: make(map[string]int),
	}
}

func (bc *Broadcast) Write(data []byte) {
	bc.m.Lock()
	defer bc.m.Unlock()
	bc.data = append(bc.data, data)
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
