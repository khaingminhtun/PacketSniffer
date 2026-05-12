package main

import (
	"log"

	httpserver "packet-sniffer/internal/api"
)

func main() {

	err := httpserver.StartServer()
	if err != nil {
		log.Fatal(err)
	}

}
