package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/aodr3w/go-tcp-feed/data"
)

var userName string

func getInput() string {
	fmt.Print(">> ")
	reader := bufio.NewReader(os.Stdin)
	name, err := reader.ReadString('\n')
	if err != nil {
		handleError(err)
	}
	return strings.TrimSpace(name)
}

func handShake(conn net.Conn) {
	for {
		fmt.Print("name (atleast 4 characters): ")
		userName = getInput()
		if len(userName) <= 3 {
			continue
		}
		break
	}
	//send the user's name to the server
	_, err := conn.Write([]byte(fmt.Sprintf("name-%s", userName)))

	if err != nil {
		handleError(err)
	}
	data := readConn(conn)
	log.Printf("%v", data.Message.Text)
}

func readConn(conn net.Conn) *data.MessagePayload {
	buf := make([]byte, 1048)
	n, err := conn.Read(buf)
	handleError(err)
	mb := buf[:n]
	msg, err := data.PayloadFromBytes(mb)
	handleError(err)
	return msg
}

func openConn(serverPort int) net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", serverPort))
	handleError(err)
	return conn
}

// responsible for connecting to server and writing new messages
func Publisher(serverPort int) {
	conn := openConn(serverPort)
	handShake(conn)
	for {
		msg := getInput()
		if len(msg) == 0 {
			continue
		}
		if strings.EqualFold(msg, "q") {
			return
		}
		mp := data.NewMessagePayload(data.NewMessage(msg, userName), 0)
		mpb, err := mp.ToBytes()
		handleError(err)
		_, err = conn.Write(mpb)
		handleError(err)
	}
}
