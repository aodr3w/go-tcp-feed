package main

import (
	"encoding/gob"
	"flag"
	"log"
	"os"

	"github.com/aodr3w/go-chat/client"
	"github.com/aodr3w/go-chat/data"
	"github.com/aodr3w/go-chat/server"
)

const writePort = 2000
const readPort = 3000

func main() {
	err := data.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	gob.Register(data.MessagePayload{})

	svr := flag.Bool("server", false, "provide to start server")
	clt := flag.Bool("publisher", false, "provide to start client side message publisher")
	cstream := flag.Bool("client-stream", false, "provide to start message stream")

	flag.Parse()

	if *svr {
		server.Start(writePort, readPort)
		os.Exit(0)
	}

	if *clt {
		client.Publisher(writePort)
		os.Exit(0)
	}

	if *cstream {
		client.StreamMessages(readPort)
		os.Exit(0)
	}

	log.Println("please provide a valid option either (server) or (client)")
}
