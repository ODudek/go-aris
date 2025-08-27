package commands

import (
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

type HSetCommand struct {
	storage storage.StorageInterface
	key     string
	field   string
	value   string
}

func (h *HSetCommand) Execute() (resp.Value, error) {
	err := h.storage.HSet(h.key, h.field, h.value)
	if err != nil {
		return resp.NewErrValue(err.Error()), err
	}
	return resp.NewIntValue(1), nil
}
