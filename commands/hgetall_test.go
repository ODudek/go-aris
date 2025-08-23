package commands

import (
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
	"testing"
)

func TestHGetAllCommand_ExistingHash(t *testing.T) {
	storage := storage.NewStorage()
	storage.HSet("hash", "field1", "value1")
	storage.HSet("hash", "field2", "value2")
	
	cmd := &HGetAllCommand{
		storage: storage,
		key:     "hash",
	}
	
	result, err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	arrVal, ok := result.(*resp.ArrValue)
	if !ok {
		t.Errorf("Expected ArrValue, got %T", result)
	}
	
	arr := arrVal.GetArray()
	if len(arr) != 4 { // 2 fields * 2 (field + value)
		t.Errorf("Expected 4 elements, got %d", len(arr))
	}
	
	fields := make(map[string]string)
	for i := 0; i < len(arr); i += 2 {
		fieldBulk, ok := arr[i].(*resp.BulkValue)
		if !ok {
			t.Errorf("Expected BulkValue for field at index %d, got %T", i, arr[i])
			continue
		}
		
		valueBulk, ok := arr[i+1].(*resp.BulkValue)
		if !ok {
			t.Errorf("Expected BulkValue for value at index %d, got %T", i+1, arr[i+1])
			continue
		}
		
		field := string(fieldBulk.GetBulkString())
		value := string(valueBulk.GetBulkString())
		fields[field] = value
	}
	
	if fields["field1"] != "value1" {
		t.Errorf("Expected field1='value1', got '%s'", fields["field1"])
	}
	
	if fields["field2"] != "value2" {
		t.Errorf("Expected field2='value2', got '%s'", fields["field2"])
	}
}

func TestHGetAllCommand_NonExistentHash(t *testing.T) {
	storage := storage.NewStorage()
	
	cmd := &HGetAllCommand{
		storage: storage,
		key:     "nonexistent",
	}
	
	result, err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	arrVal, ok := result.(*resp.ArrValue)
	if !ok {
		t.Errorf("Expected ArrValue, got %T", result)
	}
	
	arr := arrVal.GetArray()
	if len(arr) != 0 {
		t.Errorf("Expected empty array for non-existent hash, got %d elements", len(arr))
	}
}

func TestHGetAllCommand_EmptyHash(t *testing.T) {
	storage := storage.NewStorage()
	
	cmd := &HGetAllCommand{
		storage: storage,
		key:     "emptyhash",
	}
	
	result, err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	arrVal, ok := result.(*resp.ArrValue)
	if !ok {
		t.Errorf("Expected ArrValue, got %T", result)
	}
	
	arr := arrVal.GetArray()
	if len(arr) != 0 {
		t.Errorf("Expected empty array for empty hash, got %d elements", len(arr))
	}
}