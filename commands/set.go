package commands

import (
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

type SetCommand struct {
	storage storage.StorageInterface
	key     string
	value   string
}

func (s *SetCommand) Execute() (resp.Value, error) {
	err := s.storage.Set(s.key, s.value)
	if err != nil {
		return resp.NewErrValue(err.Error()), err
	}
	return resp.NewStrValue("OK"), nil
}
