package commands

import (
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

type HGetAllCommand struct {
	storage *storage.Storage
	key     string
}

func (h *HGetAllCommand) Execute() (resp.Value, error) {
	hash := h.storage.HGetAll(h.key)
	
	result := make([]resp.Value, 0, len(hash)*2)
	for field, value := range hash {
		result = append(result, resp.NewBulkValue([]byte(field)))
		result = append(result, resp.NewBulkValue([]byte(value)))
	}
	
	return resp.NewArrValue(result), nil
}