package commands

import (
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

type GetCommand struct {
	storage storage.StorageInterface
	key     string
}

func (g *GetCommand) Execute() (resp.Value, error) {
	value, exists, err := g.storage.Get(g.key)
	if err != nil {
		return resp.NewErrValue(err.Error()), err
	}
	if !exists {
		return resp.NewNullValue(), nil
	}
	return resp.NewBulkValue([]byte(value)), nil
}
