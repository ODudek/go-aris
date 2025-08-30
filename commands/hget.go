package commands

import (
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

type HGetCommand struct {
	storage storage.StorageInterface
	key     string
	field   string
}

func (h *HGetCommand) Execute() (resp.Value, error) {
	value, exists, err := h.storage.HGet(h.key, h.field)
	if err != nil {
		return resp.NewErrValue(err.Error()), err
	}
	if !exists {
		return resp.NewNullValue(), nil
	}
	return resp.NewBulkValue([]byte(value)), nil
}
