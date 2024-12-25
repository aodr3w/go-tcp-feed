package main

import (
	"encoding/gob"
	"flag"
	"log"
	"os"

	"github.com/aodr3w/go-chat/client"
	"github.com/aodr3w/go-chat/db"
	"github.com/aodr3w/go-chat/server"
)

const serverPort = 2000

func main() {
	err := db.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	gob.Register(db.Message{})

	svr := flag.Bool("server", false, "provide to start server")
	clt := flag.Bool("client", false, "provide to start client")

	flag.Parse()

	if *svr {
		dao := db.NewDAO()
		b := server.NewBroadCast(&dao)
		server.Start(serverPort, b, &dao)
		os.Exit(0)
	}
	if *clt {
		client.Start(serverPort)
		os.Exit(0)
	}

	log.Println("please provide a valid option either (server) or (client)")
}
