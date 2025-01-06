package server

import (
	"log"
	"sync"
)

func Start(writePort int, readPort int) {
	log.Println("starting server...")
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
