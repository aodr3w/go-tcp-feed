package server

import (
	"net"
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

func (bc *Broadcast) Write(conn net.Conn, data []byte) {
	addr := conn.RemoteAddr().String()
	bc.m.Lock()
	defer bc.m.Unlock()
	bc.data = append(bc.data, data)
	// Return the current length of the data slice, which represents the next index to read from.
	// This ensures the writer (publisher) can avoid reading the message it just wrote.
	// Consumers might encounter index errors if they try to read an index that is not yet written,
	// but this is expected behavior and should be handled by the caller.
	bc.readIdx[addr] = len(bc.data)
}

func (bc *Broadcast) Read(name string) []byte {
	bc.m.RLock()
	idx := bc.readIdx[name]
	defer bc.m.RUnlock()
	if idx >= 0 && idx < len(bc.data) {
		data := bc.data[idx]
		if string(data)[:len(name)] != name {
			return data
		}
	}
	return []byte{}
}
