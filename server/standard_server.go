package server

import (
	"fmt"
	"log"
	"net"

	"github.com/ODudek/go-aris/resp"
)

type StandardServer struct {
	config *Config
}

func NewStandardServer(config *Config) *StandardServer {
	return &StandardServer{
		config: config,
	}
}

func (s *StandardServer) Start() error {
	log.Printf("Starting standard server on %s", s.config.Port)
	
	l, err := net.Listen("tcp", s.config.Port)
	if err != nil {
		return err
	}
	defer l.Close()
	
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