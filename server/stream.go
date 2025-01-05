package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/aodr3w/go-chat/data"
)

func ReadMessages(port int, s *Service) {
	//read messages from the database in a stream starting from oldest to curent
	//do so continuously using an offset
	//get connection first
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		log.Println("connection received", conn.LocalAddr())
		if err != nil {
			log.Fatal(err)
		}

		go func(conn net.Conn) {
			size := 100
			offset := 0
			for {
				messages, err := s.GetMessageStream(size, offset, data.Oldest)
				if err != nil {
					log.Println("[stream-error] ", err)
					return
				}
				for _, message := range messages {
					mp := data.NewMessagePayload(message, 0)
					mpb, err := mp.ToBytes()
					if err != nil {
						log.Println("failed to serialize message due to error: ", err)
						return
					}
					_, err = conn.Write(mpb)
					if err != nil {
						log.Printf("[stream error] %v\n", err)
						return
					}
					time.Sleep(50 * time.Microsecond)
				}
				offset += len(messages)

			}
		}(conn)
	}

}
