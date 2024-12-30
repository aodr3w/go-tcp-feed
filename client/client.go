package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/aodr3w/go-chat/data"
)

func readName() string {
	fmt.Print("name: ")
	reader := bufio.NewReader(os.Stdin)
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("error reading input: %v", err)
		return ""
	}
	return name
}
func readMsg(read chan struct{}, msgChan chan string) {
	for range read {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">>")
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading input: %v", err)
			continue
		}
		msgChan <- strings.TrimSpace(msg)
	}

}

func Start(serverPort int) error {
	stop := make(chan struct{}, 1)
	inbound := make(chan *data.Message, 100)
	read := make(chan struct{}, 1)
	msgs := make(chan string, 1)

	var name string

	for {
		name = readName()
		if len(name) <= 3 {
			fmt.Println("name should be atleast 3 characters")
			continue
		}
		break
	}

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", serverPort))

	if err != nil {
		log.Fatalf("failed to reach server: %s", err)
		return err
	}

	//send the user's name to the server
	_, err = conn.Write([]byte(fmt.Sprintf("name-%s", name)))

	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	go readInput(conn, name, read, msgs)
	go readConn(conn, inbound, cancel)
	go writetoStdOut(read, inbound, stop, ctx)
	defer conn.Close()
	<-stop
	return nil

}

func readInput(conn net.Conn, name string, read chan struct{}, msgs chan string) {
	fmt.Println("enter message or q to quit")
	for {
		readMsg(read, msgs)
		txt := <-msgs
		if len(txt) == 0 {
			continue
		}
		if strings.EqualFold(txt, "q") {
			break
		}

		msg := data.Message{
			Name:      name,
			Text:      txt,
			CreatedAt: time.Now(),
		}
		msgBytes, err := msg.ToBytes()
		if err != nil {
			fmt.Printf("%s", err.Error())
		} else {
			_, err := conn.Write(msgBytes)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
func readConn(conn net.Conn, inbound chan *data.Message, cancelFunc context.CancelFunc) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			} else if opErr, ok := err.(*net.OpError); ok && strings.Contains(opErr.Err.Error(), "use of closed network connection") {
				fmt.Print("")
			} else {
				log.Printf("error reading from conn: %s\n", err)
			}
			cancelFunc()
			return
		}
		msg, err := data.FromBytes(buf[:n])
		if err != nil {
			fmt.Printf(">> serialization error: %s\n", err)
		} else {
			inbound <- msg
		}
	}
}
func writetoStdOut(read chan struct{}, inbound chan *data.Message, stop chan struct{}, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(stop)
			return
		case msg := <-inbound:
			formattedTime := msg.CreatedAt.Format("1/2/2006 15:04:05")
			fmt.Printf(">> %s [ %s - %s ]\n", msg.Text, msg.Name, formattedTime)
		default:
			read <- struct{}{}
		}
	}

}
