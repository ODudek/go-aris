package commands

import (
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

type SetCommand struct {
	storage *storage.Storage
	key     string
	value   string
}

func (s *SetCommand) Execute() (resp.Value, error) {
	s.storage.Set(s.key, s.value)
	return resp.NewStrValue("OK"), nil
}