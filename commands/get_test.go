package commands

import (
	"testing"

	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

func TestGetCommand_ExistingKey(t *testing.T) {
	storage := storage.NewStorage()
	storage.Set("testkey", "testvalue")

	cmd := &GetCommand{
		storage: storage,
		key:     "testkey",
	}

	result, err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	bulkVal, ok := result.(*resp.BulkValue)
	if !ok {
		t.Errorf("Expected BulkValue, got %T", result)
	}

	if string(bulkVal.GetBulkString()) != "testvalue" {
		t.Errorf("Expected 'testvalue', got '%s'", string(bulkVal.GetBulkString()))
	}
}

func TestGetCommand_NonExistentKey(t *testing.T) {
	storage := storage.NewStorage()

	cmd := &GetCommand{
		storage: storage,
		key:     "nonexistent",
	}

	result, err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, ok := result.(*resp.NullValue)
	if !ok {
		t.Errorf("Expected NullValue for non-existent key, got %T", result)
	}
}
