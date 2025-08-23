package commands

import (
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

type HGetCommand struct {
	storage *storage.Storage
	key     string
	field   string
}

func (h *HGetCommand) Execute() (resp.Value, error) {
	value, exists := h.storage.HGet(h.key, h.field)
	if !exists {
		return resp.NewNullValue(), nil
	}
	return resp.NewBulkValue([]byte(value)), nil
}