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
	msg, _ := reader.ReadString('\n')
	return strings.TrimSpace(msg)
}

func startClient() error {
	stop := make(chan struct{}, 1)
	var name string

	for {
		fmt.Print("name: ")
		name = readMsg()
		if len(name) <= 3 {
			fmt.Println("name should be atleast 3 characters")
			continue
		}
		break
	}

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", SERVER_PORT))

	if err != nil {
		log.Fatalf("failed to reach server: %s", err)
		return err
	}

	//send the user's name to the server
	_, err = conn.Write([]byte(fmt.Sprintf("name-%s", name)))

	if err != nil {
		return err
	}

	go readtoStdOut(conn, stop)

	fmt.Println("enter name or q to quit")
	fmt.Print(">> ")
	for {
		msg := readMsg()
		if strings.EqualFold(msg, "q") {
			break
		}
		_, err := conn.Write([]byte(msg))
		if err != nil {
			log.Fatal(err)
		}
	}
	conn.Close()
	<-stop
	return nil

}

func readtoStdOut(conn net.Conn, stop chan struct{}) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {

			if errors.Is(err, io.EOF) {
				log.Println("server closed the connection")
			} else if opErr, ok := err.(*net.OpError); ok && strings.Contains(opErr.Err.Error(), "use of closed network connection") {
				fmt.Print()
			} else {
				log.Printf("error reading from conn: %s\n", err)
			}
			close(stop)
			return
		}
		recv := string(buf[:n])
		fmt.Printf("%s\n>> ", recv)
	}
}

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
