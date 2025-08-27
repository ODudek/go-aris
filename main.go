package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ODudek/go-aris/server"
)

func main() {
	var useGnet bool
	var port string
	var clusterMode bool

	flag.BoolVar(&useGnet, "gnet", true, "Use gnet server for better performance")
	flag.StringVar(&port, "port", ":8089", "Port to listen on")
	flag.BoolVar(&clusterMode, "cluster", false, "Enable cluster mode")
	flag.Parse()

	// Override cluster mode from environment if set
	if os.Getenv("CLUSTER_ENABLED") == "true" {
		clusterMode = true
	}

	if clusterMode {
		log.Println("Starting in cluster mode")
	} else {
		log.Println("Starting in standalone mode")
	}

	config := server.NewConfig(port)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize cluster components if in cluster mode
	if config.IsClusterMode {
		if err := config.StartCluster(ctx); err != nil {
			log.Printf("Failed to start cluster components: %v", err)
			os.Exit(1)
		}
		log.Println("Cluster components initialized")
	}

	var handler server.Handler
	if useGnet {
		log.Println("Using gnet server for high performance")
		handler = server.NewGnetServer(config)
	} else {
		log.Println("Using standard Go net server")
		handler = server.NewStandardServer(config)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping server...")
		cancel()

		if err := handler.Stop(); err != nil {
			log.Printf("Error stopping server: %v", err)
		}
		os.Exit(0)
	}()

	log.Printf("Starting Aris server on port %s", port)
	if err := handler.Start(); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
