package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/aodr3w/go-chat/data"
)

/*
ReadMessages reads messages from the database continuously using an offset in ascending order
*/
func ReadMessages(port int, s *Service) {

	logger := log.New(log.Writer(), "[readMessages] ", log.LstdFlags)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		logger.Printf("error creating tcp listener: %v\n", err)
		return
	}

	logger.Printf("server is accepting connections on %d\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Printf("error accepting connection %v\n", err)
			return
		}

		logger.Println("new connection received")

		go func(conn net.Conn) {
			defer conn.Close()
			size := 100
			offset := 0
			for {
				messages, err := s.GetMessageStream(size, offset, data.Oldest)
				if err != nil {
					logger.Printf("[stream-error] %v\n", err)
					return
				}
				for _, message := range messages {
					mp := data.NewMessagePayload(message, 0)
					mpb, err := mp.ToBytes()
					if err != nil {
						logger.Printf("failed to serialize message due to error: %v\n", err)
						return
					}
					_, err = conn.Write(mpb)
					if err != nil {
						logger.Printf("[stream error] %v\n", err)
						return
					}
					time.Sleep(50 * time.Microsecond)
				}
				offset += len(messages)

			}
		}(conn)
	}

}
