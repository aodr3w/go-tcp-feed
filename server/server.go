package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/aodr3w/go-chat/data"
)

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
			log.Println("client disconnected")
		} else {
			log.Printf("error reading from conn: %s\n", err)
		}
		return
	}
	return buf[:n], nil
}

func writeConn(conn net.Conn, data []byte) {
	_, err := conn.Write(data)
	if err != nil {
		log.Printf("[writeConnError] %v\n", err)
	}
}

func handleConnection(conn net.Conn, broadcast *Broadcast, dao *data.Dao) {
	defer conn.Close()
	initial, err := readConn(conn)

	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	name, err := extractName(conn, initial)

	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	//check if name is already taken if so return an error
	user, err := dao.GetUserByName(name)
	if err != nil {
		if errors.Is(err, &data.UserNotFoundError) {
			//create new user and respond with userID that the client should save
			user, err = dao.CreateUser(name)
			if err != nil {
				writeConn(conn, []byte(err.Error()))
				return
			}
			if len(user.Name) < 1 {
				writeConn(conn, []byte("[Internal Server Error] invalid user data"))
				return
			}
			log.Printf("user successfully created: %v", user)
		} else {
			log.Printf("unknown error type: %v", err)
			writeConn(conn, []byte(err.Error()))
			return
		}
	}

	msg := data.Message{
		Name:      "system",
		Text:      fmt.Sprintf("userID-%s", user.Name),
		CreatedAt: time.Now(),
	}
	b, err := msg.ToBytes()
	if err != nil {
		log.Println("error serializing message", err.Error())
		return
	}
	conn.Write(b)

	ct := time.Now() //TODO this should be sent inside the payload
	//load messages first
	messages, err := broadcast.LoadMessages(0, 100, ct)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("error loading messages %s\n", err.Error())))
		return
	}

	for _, message := range messages {
		msgBytes, err := message.ToBytes()
		if err != nil {
			writeConn(conn, []byte(err.Error()))
			return
		}
		writeConn(conn, msgBytes)
		time.Sleep(50 * time.Millisecond)
	}
	log.Printf("user %v, ct: %v\n", user, ct)
	go func() {
		offset := 0
		size := 5
		for {
			//load 5 of the newest messages in db
			//and write them to a connection at an interval of 1 second
			latestMessages, err := dao.GetReceivedMessages(user.ID, size, offset, ct)
			if err != nil {
				writeConn(conn, []byte(err.Error()))
				return
			}
			if len(latestMessages) > 0 {
				for _, msg := range latestMessages {
					msgBytes, bytesErr := msg.ToBytes()
					if bytesErr != nil {
						writeConn(conn, []byte(bytesErr.Error()))
					} else {
						writeConn(conn, msgBytes)
					}
					time.Sleep(40 * time.Millisecond)
				}
				offset += len(latestMessages)
			}
		}

	}()

	for {
		//handles reading messages from connection and publishing them to broadcast queue
		recv, err := readConn(conn)
		if err != nil {
			log.Printf("%v\n", err)
			return
		}
		err = broadcast.Write(recv)
		if err != nil {
			writeConn(conn, []byte(err.Error()))
			return
		}
	}
}

func Start(SERVER_PORT int, broadcast *Broadcast, dao *data.Dao) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", SERVER_PORT))
	if err != nil {
		return err
	}
	defer ln.Close()
	log.Printf("server is accepting connections on %d\n", SERVER_PORT)
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("connection closed")
				continue
			}
			return err
		}
		go handleConnection(conn, broadcast, dao)
	}
}
