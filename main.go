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

const serverPort = 2000
const streamPort = 3000

func main() {
	err := data.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	gob.Register(data.MessagePayload{})

	svr := flag.Bool("server", false, "provide to start server")
	clt := flag.Bool("client", false, "provide to start client")
	cstream := flag.Bool("client-stream", false, "provide to start message stream")
	sstream := flag.Bool("server-stream", false, "start server side message stream on port 3000")

	flag.Parse()

	if *svr {
		dao := data.NewDAO()
		b := server.NewBroadCast(&dao)
		server.Start(serverPort, b, &dao)
		os.Exit(0)
	}
	if *clt {
		client.Start(serverPort)
		os.Exit(0)
	}
	if *cstream {
		client.StreamMessages(streamPort)
		os.Exit(0)
	}
	if *sstream {
		dao := data.NewDAO()
		server.StreamMessages(streamPort, &dao)
		os.Exit(0)
	}

	log.Println("please provide a valid option either (server) or (client)")
}
