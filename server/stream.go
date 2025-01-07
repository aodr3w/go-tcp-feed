package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/aodr3w/go-chat/data"
)

/*
/*
ReadMessages starts a TCP server on the specified port to stream messages to connected clients.

This function:
- Listens for incoming TCP connections on the specified port.
- For each new connection, starts a goroutine to stream messages to the client.
- Reads messages from the database using a pagination-like approach (with `size` and `offset`).
- Serializes each message into bytes and sends it over the TCP connection.
- Handles errors gracefully, including issues with database access, message serialization, or client disconnection.

Parameters:
- port: The port number on which the server listens for incoming TCP connections.
- s: A pointer to the Service layer, which provides methods to fetch and process messages from the database.

Errors are logged using a logger with the prefix `[readMessages]`. Each client connection runs in its own goroutine, and the server continues to accept new connections until an error occurs or the server is stopped.
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
