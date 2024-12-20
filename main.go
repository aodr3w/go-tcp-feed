package main

import (
	"flag"
	"log"
	"os"

	"github.com/aodr3w/go-chat/client"
	"github.com/aodr3w/go-chat/server"
)

const serverPort = 2000

func main() {
	svr := flag.Bool("server", false, "provide to start server")
	clt := flag.Bool("client", false, "provide to start client")

	flag.Parse()

	if *svr {
		b := server.NewBroadCast()
		server.Start(serverPort, b)
		os.Exit(0)
	}
	if *clt {
		client.Start(serverPort)
		os.Exit(0)
	}

	log.Println("please provide a valid option either (server) or (client)")
}
