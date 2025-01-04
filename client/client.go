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

var userName string

func printMessage(msg *data.Message) {
	var pref string
	if msg.Name == userName {
		pref = "<<"
	} else {
		pref = ">>"
	}
	formattedTime := msg.CreatedAt.Format("1/2/2006 15:04:05")
	fmt.Printf("%s %s [ %s - %s ]\n", pref, msg.Text, msg.Name, formattedTime)
}
func readStdIn() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	txt, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %v", err)
	}
	return txt, nil
}
func readName() string {
	fmt.Print("name: ")
	name, err := readStdIn()
	if err != nil {
		log.Println(err)
		return ""
	}
	return name
}

func readMsg() string {
	fmt.Print("<< ")
	msg, err := readStdIn()
	if err != nil {
		log.Printf("error reading input: %v", err)
		return ""
	}
	return strings.TrimSpace(msg)
}

func Start(serverPort int) error {
	startedAt := time.Now()
	//channel for messages from current session
	inboundChan := make(chan *data.MessagePayload, 100)
	//channel to signal a message can be read from the input
	readChan := make(chan struct{}, 1)
	//channel for messages before current session
	historyChan := make(chan *data.MessagePayload, 100)
	historyDone := make(chan struct{}, 1)
	connDataChan := make(chan []byte)
	connErrChan := make(chan error)
	exitChan := make(chan struct{}, 1)

	fmt.Println("TIP: ENTER Q to quit")

	for {
		userName = readName()
		if len(userName) <= 3 {
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
	_, err = conn.Write([]byte(fmt.Sprintf("name-%s", userName)))

	if err != nil {
		return err
	}

	appCtx, appCancel := context.WithCancel(context.Background())

	go loadHistory(historyChan, historyDone)
	go readInput(conn, userName, readChan, exitChan)
	go handleConn(connDataChan, connErrChan, inboundChan, historyChan, historyDone, appCtx, startedAt)
	go readConn(conn, connDataChan, connErrChan)
	go writeSessionMessages(inboundChan, appCtx, readChan, historyDone)
	<-exitChan
	appCancel()
	conn.Close()
	return nil

}

func loadHistory(historyChan chan *data.MessagePayload, historyDone chan struct{}) {
	//loads messages created before the current sessions startedAt time
	//should be remotely cancelled once the writeSessionMessages is called

	var count int
	for msg := range historyChan {
		printMessage(&msg.Message)
		count += 1
		if count >= msg.Count {
			historyDone <- struct{}{}
			log.Println("load history done.")
			return
		}
	}
}

func readInput(conn net.Conn, name string, readChan chan struct{}, exitChan chan struct{}) {
	for range readChan {
		txt := readMsg()
		if len(txt) == 0 {
			continue
		}
		if strings.EqualFold(txt, "q") {
			exitChan <- struct{}{}
			return
		}

		payload := data.MessagePayload{
			Message: data.Message{
				Name:      name,
				Text:      txt,
				CreatedAt: time.Now(),
			},
		}
		msgBytes, err := payload.ToBytes()
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

func readConn(conn net.Conn, dataChan chan<- []byte, errChan chan<- error) {
	defer close(dataChan)
	defer close(errChan)
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			errChan <- err
			return
		}
		dataChan <- buf[:n]
	}

}
func handleConn(
	connDataChan chan []byte,
	connErrChan chan error,
	inboundChan chan *data.MessagePayload,
	historyChan chan *data.MessagePayload,
	historyDone chan struct{},
	appCtx context.Context,
	sessionStartTime time.Time) {

	for {
		select {
		case <-appCtx.Done():
			return
		case connData, ok := <-connDataChan:
			if !ok {
				return
			}
			msg, err := data.PayloadFromBytes(connData)
			if err != nil {
				fmt.Printf(">> serialization error: %s\n", err)
			} else {
				if strings.Contains(msg.Text, "userID-") {
					if msg.Count == 0 {
						close(historyChan)
						historyDone <- struct{}{}
					}
				} else if strings.Contains(msg.Text, "system") {
					printMessage(&msg.Message)
				} else if sessionStartTime.After(msg.CreatedAt) {
					historyChan <- msg
				} else {
					inboundChan <- msg
				}

			}
		case err, ok := <-connErrChan:
			if !ok {
				return
			}
			log.Printf("Error reading from conn %v\n", err)
			return
		}

	}
}
func writeSessionMessages(inboundChan chan *data.MessagePayload,
	appCtx context.Context,
	readChan chan struct{}, historyDone chan struct{},
) {
	//cancel the history goroutine before starting current session
	<-historyDone
	var count int
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-appCtx.Done():
			return
		case msg := <-inboundChan:
			count += 1
			printMessage(&msg.Message)
			readChan <- struct{}{}
		case <-ticker.C:
			readChan <- struct{}{}
		}
	}

}
