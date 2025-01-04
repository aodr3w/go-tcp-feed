package client

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/aodr3w/go-chat/data"
)

// connects to the server and writes all messages from earlies to latest in near real time

func printStreamMessage(msg *data.Message) {
	txt := strings.TrimSpace(msg.Text)
	name := strings.TrimSpace(msg.Name)
	fmt.Printf("%s >> %s\n", name, txt)
}

func handleError(err error) {
	if err != nil {
		log.Fatalf("[stream-error] %v", err)
	}

}
func StreamMessages(streamPort int) {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", streamPort))
	if err != nil {
		handleError(err)
	}

	buff := make([]byte, 1024)

	for {
		n, err := conn.Read(buff)
		handleError(err)
		msg, err := data.PayloadFromBytes(buff[:n])
		handleError(err)
		printStreamMessage(&msg.Message)
	}

}
