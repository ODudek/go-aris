package storage

import "testing"

func TestStorage_SetAndGet(t *testing.T) {
	s := NewStorage()

	s.Set("key1", "value1")

	value, exists, err := s.Get("key1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !exists {
		t.Error("Expected key to exist")
	}

	if value != "value1" {
		t.Errorf("Expected 'value1', got '%s'", value)
	}

	_, exists, err = s.Get("nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if exists {
		t.Error("Expected key to not exist")
	}
}

func TestStorage_HSetAndHGet(t *testing.T) {
	s := NewStorage()

	s.HSet("hash1", "field1", "value1")

	value, exists, err := s.HGet("hash1", "field1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !exists {
		t.Error("Expected hash field to exist")
	}

	if value != "value1" {
		t.Errorf("Expected 'value1', got '%s'", value)
	}

	_, exists, err = s.HGet("hash1", "nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if exists {
		t.Error("Expected hash field to not exist")
	}

	_, exists, err = s.HGet("nonexistent", "field1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if exists {
		t.Error("Expected hash to not exist")
	}
}

func TestStorage_HGetAll(t *testing.T) {
	s := NewStorage()

	s.HSet("hash1", "field1", "value1")
	s.HSet("hash1", "field2", "value2")

	hash, err := s.HGetAll("hash1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(hash) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(hash))
	}

	if hash["field1"] != "value1" {
		t.Errorf("Expected 'value1', got '%s'", hash["field1"])
	}

	if hash["field2"] != "value2" {
		t.Errorf("Expected 'value2', got '%s'", hash["field2"])
	}

	emptyHash, err := s.HGetAll("nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(emptyHash) != 0 {
		t.Errorf("Expected empty hash, got %d fields", len(emptyHash))
	}
}
