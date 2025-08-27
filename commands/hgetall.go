package commands

import (
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

type HGetAllCommand struct {
	storage storage.StorageInterface
	key     string
}

func (h *HGetAllCommand) Execute() (resp.Value, error) {
	fields, err := h.storage.HGetAll(h.key)
	if err != nil {
		return resp.NewErrValue(err.Error()), err
	}

	values := make([]resp.Value, 0, len(fields)*2)
	for field, value := range fields {
		values = append(values, resp.NewBulkValue([]byte(field)))
		values = append(values, resp.NewBulkValue([]byte(value)))
	}

	return resp.NewArrValue(values), nil
}
