package storage

import (
	"sync"
)

// StorageInterface defines the interface for storage operations
type StorageInterface interface {
	Set(key, value string) error
	Get(key string) (string, bool, error)
	HSet(key, field, value string) error
	HGet(key, field string) (string, bool, error)
	HGetAll(key string) (map[string]string, error)
}

type Storage struct {
	mu     sync.RWMutex
	data   map[string]string
	hashes map[string]map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		data:   make(map[string]string),
		hashes: make(map[string]map[string]string),
	}
}

func (s *Storage) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *Storage) Get(key string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	return value, exists, nil
}

func (s *Storage) HSet(key, field, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.hashes[key] == nil {
		s.hashes[key] = make(map[string]string)
	}
	s.hashes[key][field] = value
	return nil
}

func (s *Storage) HGet(key, field string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hash, exists := s.hashes[key]
	if !exists {
		return "", false, nil
	}

	value, exists := hash[field]
	return value, exists, nil
}

func (s *Storage) HGetAll(key string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hash, exists := s.hashes[key]
	if !exists {
		return make(map[string]string), nil
	}

	result := make(map[string]string)
	for field, value := range hash {
		result[field] = value
	}
	return result, nil
}
