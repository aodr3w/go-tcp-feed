package server

import (
	"fmt"
	"log"
	"sync"
)

func NewLogger(prefix string) *log.Logger {
	return log.New(
		log.Writer(),
		fmt.Sprintf("%s ", prefix),
		log.LstdFlags,
	)
}

func Start(writePort int, readPort int) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		s := NewService()
		WriteMessages(writePort, s)
		defer wg.Done()
	}()

	wg.Add(1)
	go func() {
		s := NewService()
		ReadMessages(readPort, s)
		wg.Done()
	}()
	wg.Wait()
}
