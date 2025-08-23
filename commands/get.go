package commands

import (
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

type GetCommand struct {
	storage *storage.Storage
	key     string
}

func (g *GetCommand) Execute() (resp.Value, error) {
	value, exists := g.storage.Get(g.key)
	if !exists {
		return resp.NewNullValue(), nil
	}
	return resp.NewBulkValue([]byte(value)), nil
}