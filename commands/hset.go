package commands

import (
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

type HSetCommand struct {
	storage *storage.Storage
	key     string
	field   string
	value   string
}

func (h *HSetCommand) Execute() (resp.Value, error) {
	h.storage.HSet(h.key, h.field, h.value)
	return resp.NewIntValue(1), nil
}
