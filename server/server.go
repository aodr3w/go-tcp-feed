package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/aodr3w/go-chat/db"
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
		log.Println(fmt.Sprintf("[writeConnError] %v", err))
	}
}

func handleConnection(conn net.Conn, broadcast *Broadcast, dao *db.Dao) {
	defer conn.Close()
	//read latest message from broadcast queue
	//TODO add a readALL function that loads all messages from earliest to last when connection is first established
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
	existingUser, err := dao.GetUserByName(name)

	if err != nil {
		if errors.Is(err, &db.UserNotFoundError) {
			//create new user and respond with userID that the client should save
			newUser, err := dao.CreateUser(name)
			if err != nil {
				writeConn(conn, []byte(err.Error()))
				return
			}
			if len(newUser.Name) < 1 {
				writeConn(conn, []byte(fmt.Sprintf("[Internal Server Error] invalid user data")))
				return
			}

			conn.Write([]byte(fmt.Sprintf("userID-%s", newUser.Name)))
		} else {
			writeConn(conn, []byte(err.Error()))
			return

		}
	} else {
		conn.Write([]byte(fmt.Sprintf("userID-%s", existingUser.Name)))
	}
	//load messages first
	messages, err := broadcast.LoadMessages(0, 100)

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
	}

	go func() {
		offset := 0
		size := 5
		for {
			//load 5 of the newest messages in db
			//and write them to a connection at an interval of 1 second
			messages, err := dao.GetMessages(size, offset, db.Latest)
			if err != nil {
				writeConn(conn, []byte(err.Error()))
				return
			}
			for _, msg := range messages {
				msgBytes, bytesErr := msg.ToBytes()
				if bytesErr != nil {
					writeConn(conn, []byte(bytesErr.Error()))
				} else {
					writeConn(conn, msgBytes)
				}
				time.Sleep(time.Second * 1)
			}
			offset += size
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

func Start(SERVER_PORT int, broadcast *Broadcast, dao *db.Dao) error {
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
