package storage

import (
	"sync"
)

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

func (s *Storage) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Storage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	return value, exists
}

func (s *Storage) HSet(key, field, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.hashes[key] == nil {
		s.hashes[key] = make(map[string]string)
	}
	s.hashes[key][field] = value
}

func (s *Storage) HGet(key, field string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	hash, exists := s.hashes[key]
	if !exists {
		return "", false
	}
	
	value, exists := hash[field]
	return value, exists
}

func (s *Storage) HGetAll(key string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	hash, exists := s.hashes[key]
	if !exists {
		return make(map[string]string)
	}
	
	result := make(map[string]string)
	for field, value := range hash {
		result[field] = value
	}
	return result
}