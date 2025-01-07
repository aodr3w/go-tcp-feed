package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/aodr3w/go-chat/data"
)

var (
	logger = NewLogger("[writeMessages] ")
)

/*
/*
handleConnection manages an individual client connection.
It performs the following steps:
1. Reads the initial handshake data from the client.
2. Extracts the client's name or assigns a default identifier if the name is missing.
3. Ensures the client's name is unique or creates a new user in the database.
4. Sends an initial system message to the client.
5. Continuously reads messages from the client and writes them to the database.

Parameters:
- conn: The active TCP connection with the client.
- service: The Service layer for interacting with the database.

The connection is closed upon completion or in case of an error.
*/
func WriteMessages(SERVER_PORT int, s *Service) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", SERVER_PORT))
	if err != nil {
		logger.Printf("error creating listener: %v\n", err)
		return
	}
	defer ln.Close()
	logger.Printf("server is accepting connections on %d\n", SERVER_PORT)
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				logger.Println("connection closed")
				continue
			}
			logger.Printf("error accepting new connection %v", err)
			return
		}
		go handleConnection(conn, s)
	}
}

func extractName(conn net.Conn, data []byte) (string, error) {
	//extract name or return remoteAddr
	v := string(data)
	if strings.Contains(v, "name-") {
		ss := strings.Split(v, "name-")
		return ss[len(ss)-1], nil
	}
	return "", fmt.Errorf("name prefix not found in conn from address %s", conn.RemoteAddr().String())
}

func readConn(conn net.Conn) (data []byte, err error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			logger.Println("client disconnected")
		} else {
			logger.Printf("error reading from conn: %s\n", err)
		}
		return
	}
	return buf[:n], nil
}

func writeConn(conn net.Conn, data []byte) {
	_, err := conn.Write(data)
	if err != nil {
		logger.Printf("%v\n", err)
	}
}

/*
handleConnection handles the handshake step with clients
and writes messages to the database
*/
func handleConnection(conn net.Conn, service *Service) {
	defer conn.Close()
	initial, err := readConn(conn)

	if err != nil {
		logger.Printf("%v\n", err)
		return
	}

	name, err := extractName(conn, initial)

	if err != nil {
		logger.Printf("%v\n", err)
		return
	}
	//check if name is already taken if so return an error
	user, err := service.GetUserByName(name)
	if err != nil {
		if errors.Is(err, &data.UserNotFoundError) {
			//create new user and respond with userID that the client should save
			user, err = service.CreateUser(name)
			if err != nil {
				writeConn(conn, []byte(err.Error()))
				return
			}
			if len(user.Name) < 1 {
				writeConn(conn, []byte("[Internal Server Error] invalid user data"))
				return
			}
			logger.Printf("user successfully created: %v", user)
		} else {
			logger.Printf("unknown error type: %v", err)
			writeConn(conn, []byte(err.Error()))
			return
		}
	}

	msgCount, err := service.GetMessageCount()
	if err != nil {
		logger.Println("error getting message count", err.Error())
		return
	}
	payload := data.MessagePayload{
		Count: msgCount,
		Message: data.Message{
			Name:      "system",
			Text:      fmt.Sprintf("userID-%s", user.Name),
			CreatedAt: time.Now(),
		},
	}
	b, err := payload.ToBytes()
	if err != nil {
		logger.Println("error serializing message", err.Error())
		return
	}

	_, err = conn.Write(b)

	if err != nil {
		logger.Printf("error writing system message: %v\n", err)
	}

	//continuosly write new messages from the connection to
	for {
		//handles reading messages from connection and writing them to the database via service Layer.
		recv, err := readConn(conn)
		if err != nil {
			logger.Printf("%v\n", err)
			return
		}
		err = service.Write(recv)
		if err != nil {
			writeConn(conn, []byte(err.Error()))
			return
		}
	}
}
