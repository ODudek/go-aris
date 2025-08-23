package server

import (
	"github.com/ODudek/go-aris/commands"
	"github.com/ODudek/go-aris/storage"
)

type Handler interface {
	Start() error
	Stop() error
}

type Config struct {
	Port    string
	Storage *storage.Storage
	Parser  *commands.Parser
}

func NewConfig(port string) *Config {
	storage := storage.NewStorage()
	parser := commands.NewParser(storage)

	return &Config{
		Port:    port,
		Storage: storage,
		Parser:  parser,
	}
}
