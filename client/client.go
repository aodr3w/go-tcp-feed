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
	"sync"
	"time"

	"github.com/aodr3w/go-chat/data"
)

func printMessage(msg *data.Message) {
	formattedTime := msg.CreatedAt.Format("1/2/2006 15:04:05")
	fmt.Printf(">> %s [ %s - %s ]\n", msg.Text, msg.Name, formattedTime)
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
	fmt.Print(">>")
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
	inboundChan := make(chan *data.Message, 100)
	//channel to signal a message can be read from the input
	readChan := make(chan struct{}, 1)
	//channel for messages before current session
	historyChan := make(chan *data.Message, 100)

	historyReady := make(chan struct{}, 1)
	sessionReady := make(chan struct{}, 1)

	var name string

	fmt.Println("TIP: ENTER Q to quit")

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

	appWg := &sync.WaitGroup{}
	appCtx, appCancel := context.WithCancel(context.Background())

	go loadHistory(appCtx, historyChan, readChan, historyReady)
	go readInput(conn, name, readChan, appCancel)
	appWg.Add(1)
	go readConn(conn, inboundChan, historyChan, historyReady, sessionReady, appCtx, appWg, startedAt)
	appWg.Add(1)
	go writeSessionMessages(inboundChan, appCtx, appWg, readChan, sessionReady)
	defer conn.Close()
	defer appWg.Wait()
	return nil

}

func loadHistory(ctx context.Context, historyChan chan *data.Message, readChan chan struct{}, ready chan struct{}) {
	//loads messages created before the current sessions startedAt time
	//should be remotely cancelled once the writeSessionMessages is called
	var sentReady bool
	for {
		select {
		case <-ctx.Done():
			close(historyChan)
			readChan <- struct{}{}
			return
		case msg := <-historyChan:
			printMessage(msg)
		default:
			if !sentReady {
				ready <- struct{}{}
				sentReady = true
			}
		}
	}
}
func readInput(conn net.Conn, name string, readChan chan struct{}, appCancel context.CancelFunc) {
	for {
		<-readChan
		txt := readMsg()
		if len(txt) == 0 {
			continue
		}
		if strings.EqualFold(txt, "q") {
			appCancel()
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
func readConn(
	conn net.Conn,
	inboundChan chan *data.Message,
	historyChan chan *data.Message,
	historyReady chan struct{},
	sessionReady chan struct{},
	appCtx context.Context,
	appWg *sync.WaitGroup,
	sessionStartTime time.Time) {
	defer appWg.Done()

	<-historyReady
	<-sessionReady

	for {
		select {
		case <-appCtx.Done():
			log.Println("[client] stopping readConn function")
			return
		default:
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
				return
			}
			msg, err := data.FromBytes(buf[:n])
			if err != nil {
				fmt.Printf(">> serialization error: %s\n", err)
			} else {
				//if the message is before session start time, its loading history
				//otherwise its from the current chat session
				if sessionStartTime.After(msg.CreatedAt) {
					historyChan <- msg
				} else {
					inboundChan <- msg
				}

			}
		}

	}
}
func writeSessionMessages(inboundChan chan *data.Message,
	appCtx context.Context, appWg *sync.WaitGroup,
	readChan chan struct{},
	sessionReady chan struct{},
) {
	//cancel the history goroutine before starting current session
	var count int
	var sessionReadySent bool
	defer appWg.Done()
	for {
		select {
		case <-appCtx.Done():
			log.Println("[client] stopping writeSessionMessages")
			return
		case msg := <-inboundChan:
			count += 1
			printMessage(msg)
			readChan <- struct{}{}
		default:
			if !sessionReadySent {
				sessionReady <- struct{}{}
				sessionReadySent = true
			}
		}
	}

}
