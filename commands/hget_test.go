package commands

import (
	"testing"

	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

func TestHGetCommand_ExistingField(t *testing.T) {
	storage := storage.NewStorage()
	storage.HSet("hash", "field1", "value1")

	cmd := &HGetCommand{
		storage: storage,
		key:     "hash",
		field:   "field1",
	}

	result, err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	bulkVal, ok := result.(*resp.BulkValue)
	if !ok {
		t.Errorf("Expected BulkValue, got %T", result)
	}

	if string(bulkVal.GetBulkString()) != "value1" {
		t.Errorf("Expected 'value1', got '%s'", string(bulkVal.GetBulkString()))
	}
}

func TestHGetCommand_NonExistentField(t *testing.T) {
	storage := storage.NewStorage()
	storage.HSet("hash", "field1", "value1")

	cmd := &HGetCommand{
		storage: storage,
		key:     "hash",
		field:   "nonexistent",
	}

	result, err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, ok := result.(*resp.NullValue)
	if !ok {
		t.Errorf("Expected NullValue for non-existent field, got %T", result)
	}
}

func TestHGetCommand_NonExistentHash(t *testing.T) {
	storage := storage.NewStorage()

	cmd := &HGetCommand{
		storage: storage,
		key:     "nonexistent",
		field:   "field1",
	}

	result, err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, ok := result.(*resp.NullValue)
	if !ok {
		t.Errorf("Expected NullValue for non-existent hash, got %T", result)
	}
}
