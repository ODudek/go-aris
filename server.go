package main

import (
	"fmt"
	"github.com/ODudek/go-aris/resp"
	"net"
)

type Server struct {
	port string
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	defer l.Close()
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	handleData(conn)
	return nil
}

func handleData(conn net.Conn) error {
	for {
		r := resp.NewReader(conn)
		//w := resp.NewWriter(conn)
		v, err := r.Read()
		if err != nil {
			return err
		}
		fmt.Println(v.GetValue())
		conn.Write([]byte("+OK\r\n"))
		//w.Write(v)

		//buf := make([]byte, 1024)
		//_, err := conn.Read(buf)
		//if err != nil {
		//	if err == io.EOF {
		//		break
		//	}
		//	return err
		//}
		//conn.Write([]byte("+OK\r\n"))
	}
}

func NewServer(port string) *Server {
	return &Server{
		port: port,
	}
}
