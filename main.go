package main

import (
	"flag"
	"log"
	"os"

	"github.com/ODudek/go-aris/server"
)

func main() {
	var useGnet bool
	var port string

	flag.BoolVar(&useGnet, "gnet", true, "Use gnet server for better performance")
	flag.StringVar(&port, "port", ":8089", "Port to listen on")
	flag.Parse()

	config := server.NewConfig(port)

	var handler server.Handler
	if useGnet {
		log.Println("Using gnet server for high performance")
		handler = server.NewGnetServer(config)
	} else {
		log.Println("Using standard Go net server")
		handler = server.NewStandardServer(config)
	}

	if err := handler.Start(); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
