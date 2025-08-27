package commands

import (
	"testing"

	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
)

func TestSetCommand(t *testing.T) {
	storage := storage.NewStorage()

	cmd := &SetCommand{
		storage: storage,
		key:     "testkey",
		value:   "testvalue",
	}

	result, err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	strVal, ok := result.(*resp.StrValue)
	if !ok {
		t.Errorf("Expected StrValue, got %T", result)
	}

	if strVal.GetString() != "OK" {
		t.Errorf("Expected 'OK', got '%s'", strVal.GetString())
	}

	value, exists, err := storage.Get("testkey")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !exists {
		t.Error("Expected key to exist in storage after SET")
	}

	if value != "testvalue" {
		t.Errorf("Expected 'testvalue' in storage, got '%s'", value)
	}
}
