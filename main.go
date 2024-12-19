package main

import (
	"flag"
	"log"
	"os"
)

const serverPort = 2000

func main() {
	server := flag.Bool("server", false, "provide to start server")
	client := flag.Bool("client", false, "provide to start client")

	flag.Parse()

	if *server {
		startServer(serverPort)
		os.Exit(0)
	}
	if *client {
		startClient(serverPort)
		os.Exit(0)
	}

	log.Println("please provide a valid option either (server) or (client)")
}
