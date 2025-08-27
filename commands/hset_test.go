package commands

import (
	"testing"

	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

func TestHSetCommand(t *testing.T) {
	storage := storage.NewStorage()

	cmd := &HSetCommand{
		storage: storage,
		key:     "hash",
		field:   "field1",
		value:   "value1",
	}

	result, err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	intVal, ok := result.(*resp.IntValue)
	if !ok {
		t.Errorf("Expected IntValue, got %T", result)
	}

	expected := ":1\r\n"
	if string(intVal.GetValue()) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(intVal.GetValue()))
	}

	value, exists, err := storage.HGet("hash", "field1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !exists {
		t.Error("Expected hash field to exist in storage after HSET")
	}

	if value != "value1" {
		t.Errorf("Expected 'value1' in storage, got '%s'", value)
	}
}
