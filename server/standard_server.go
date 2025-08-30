package server

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/ODudek/go-aris/resp"
)

type StandardServer struct {
	config     *Config
	httpServer *http.Server
	listener   net.Listener
}

func NewStandardServer(config *Config) *StandardServer {
	return &StandardServer{
		config: config,
	}
}

func (s *StandardServer) Start() error {
	log.Printf("Starting standard server on %s", s.config.Port)

	// Setup minimal HTTP server for health checks only
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.config.HealthCheckHandler)

	// Start HTTP server on a different port for health checks
	go func() {
		healthPort := ":8080" // Default health check port
		healthServer := &http.Server{
			Addr:    healthPort,
			Handler: mux,
		}
		log.Printf("Starting health check server on %s", healthPort)
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health check server error: %v", err)
		}
	}()

	l, err := net.Listen("tcp", s.config.Port)
	if err != nil {
		return err
	}
	s.listener = l

	// Handle Redis protocol connections
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *StandardServer) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *StandardServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("New connection from %s", conn.RemoteAddr())

	for {
		r := resp.NewReader(conn)
		w := resp.NewWriter(conn)

		v, err := r.Read()
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}

		fmt.Printf("Received: %s\n", string(v.GetValue()))

		cmd, err := s.config.Parser.Parse(v)
		if err != nil {
			errorResp := resp.NewErrValue(err.Error())
			w.Write(errorResp)
			continue
		}

		result, err := cmd.Execute()
		if err != nil {
			errorResp := resp.NewErrValue(err.Error())
			w.Write(errorResp)
			continue
		}

		w.Write(result)
	}
}
