package server

import (
	"bytes"
	"fmt"
	"log"

	"github.com/ODudek/go-aris/resp"
	"github.com/panjf2000/gnet/v2"
)

type GnetServer struct {
	*gnet.BuiltinEventEngine
	config *Config
	addr   string
	engine gnet.Engine
}

func NewGnetServer(config *Config) *GnetServer {
	return &GnetServer{
		BuiltinEventEngine: &gnet.BuiltinEventEngine{},
		config:             config,
		addr:               "tcp://localhost" + config.Port,
	}
}

func (s *GnetServer) Start() error {
	log.Printf("Starting gnet server on %s", s.config.Port)
	return gnet.Run(s, s.addr, gnet.WithMulticore(true))
}

func (s *GnetServer) Stop() error {
	return nil
}

func (s *GnetServer) OnBoot(engine gnet.Engine) gnet.Action {
	s.engine = engine
	log.Printf("Gnet server started on %s", s.addr)
	return gnet.None
}

func (s *GnetServer) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	log.Printf("New connection from %s", c.RemoteAddr())
	return nil, gnet.None
}

func (s *GnetServer) OnClose(c gnet.Conn, err error) gnet.Action {
	log.Printf("Connection closed from %s", c.RemoteAddr())
	return gnet.None
}

func (s *GnetServer) OnTraffic(c gnet.Conn) gnet.Action {
	buf := make([]byte, c.InboundBuffered())
	n, err := c.Read(buf)
	if err != nil {
		log.Printf("Error reading from connection: %v", err)
		return gnet.Close
	}

	if n == 0 {
		return gnet.None
	}

	buf = buf[:n]

	reader := resp.NewReader(bytes.NewReader(buf))
	value, err := reader.Read()
	if err != nil {
		log.Printf("Error parsing RESP: %v", err)
		errorResp := resp.NewErrValue(err.Error())
		c.Write(errorResp.GetValue())
		return gnet.None
	}

	fmt.Printf("Received: %s\n", string(value.GetValue()))

	cmd, err := s.config.Parser.Parse(value)
	if err != nil {
		errorResp := resp.NewErrValue(err.Error())
		c.Write(errorResp.GetValue())
		return gnet.None
	}

	result, err := cmd.Execute()
	if err != nil {
		errorResp := resp.NewErrValue(err.Error())
		c.Write(errorResp.GetValue())
		return gnet.None
	}

	c.Write(result.GetValue())
	return gnet.None
}
