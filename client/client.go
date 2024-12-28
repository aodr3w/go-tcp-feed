package client

import (
	"bufio"
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

func readMsg() string {
	reader := bufio.NewReader(os.Stdin)

	msg, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("error reading input: %v", err)
		return ""
	}

	return strings.TrimSpace(msg)
}

func Start(serverPort int) error {
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

	go readtoStdOut(conn, stop)

	fmt.Println("enter message or q to quit")
	for {
		txt := readMsg()
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
				continue
			} else if opErr, ok := err.(*net.OpError); ok && strings.Contains(opErr.Err.Error(), "use of closed network connection") {
				fmt.Print("")
			} else {
				log.Printf("error reading from conn: %s\n", err)
			}
			close(stop)
			return
		}
		msg, err := data.FromBytes(buf[:n])
		if err != nil {
			fmt.Printf(">> serialization error: %s\n", err)
		} else {
			formattedTime := msg.CreatedAt.Format("1/2/2006 15:04:05")
			fmt.Printf(">> %s [ %s - %s ]\n", msg.Text, msg.Name, formattedTime)
		}
	}
}
