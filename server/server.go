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
	_, err = dao.GetUserByName(name)

	//TODO Registering a USER should be seperate logic
	if err != nil {
		log.Printf("[getUserError]  %v\n", err)
		conn.Write([]byte(err.Error()))
		return
	}

	go func() {
		for {
			latestMsg := broadcast.Read(name)
			//write latest message from broadcast for this connection
			//back to the client
			if len(latestMsg) > 0 {
				//write the message the client sent to the br
				_, writeErr := conn.Write(latestMsg)
				if writeErr != nil {
					log.Printf("error writing to conn: %s\n", writeErr)
					return
				}
			}
			time.Sleep(1 * time.Second)

		}
	}()

	for {
		//handles reading messages from connection and publishing them to broadcast queue
		recv, err := readConn(conn)
		if err != nil {
			log.Printf("%v\n", err)
			return
		}
		broadcast.Write([]byte(fmt.Sprintf("%s: %s", name, recv)))
		log.Printf("message %s pushed\n", string(recv))
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
