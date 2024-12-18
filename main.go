package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const SERVER_PORT = 2000

func readMsg() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(">> ")
	msg, _ := reader.ReadString('\n')
	return strings.TrimSpace(msg)
}

func startClient() error {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", SERVER_PORT))

	if err != nil {
		return err
	}

	defer conn.Close()

	fmt.Println("enter message or q to quit")
	for {
		msg := readMsg()
		if strings.EqualFold(msg, "q") {
			log.Println("exiting..")
			break
		}
		_, err := conn.Write([]byte(msg))
		if err != nil {
			log.Fatal(err)
		}

		resp, err := io.ReadAll(conn)

		if err != nil {
			log.Fatal(err)
		}
		log.Printf("server response: %s\n", string(resp))
	}
	return nil

}

func handleConnection(conn net.Conn) {
	//read the data off the connection and echo it back
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
		resp := fmt.Sprintf("echo: %s", recv)
		_, writeErr := conn.Write([]byte(resp))
		if writeErr != nil {
			log.Printf("error writing to conn: %s\n", writeErr)
			return
		}
	}
}

func startServer() error {
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

func main() {
	server := flag.Bool("server", false, "provide to start server")
	client := flag.Bool("client", false, "provide to start client")

	flag.Parse()

	if *server {
		startServer()
		os.Exit(0)
	}
	if *client {
		startClient()
		os.Exit(0)
	}

	log.Println("please provide a valid option either (server) or (client)")
}
