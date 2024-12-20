package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func handleConnection(conn net.Conn) {
	//read the data off the connection and echo it back
	var name string
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("client disconnected")
				return
			}
			log.Printf("error reading from conn: %s\n", err)
			return
		}
		recv := string(buf[:n])
		if strings.Contains(recv, "name-") {
			ss := strings.Split(recv, "name-")
			name = ss[len(ss)-1]
		} else {
			resp := fmt.Sprintf("%s: %s", name, recv)
			_, writeErr := conn.Write([]byte(resp))
			if writeErr != nil {
				log.Printf("error writing to conn: %s\n", writeErr)
				return
			}
		}

	}
}

func Start(SERVER_PORT int, b *Broadcast) error {
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
		go handleConnection(conn)
	}
}
