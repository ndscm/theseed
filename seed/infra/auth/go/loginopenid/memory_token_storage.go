package loginopenid

import (
	"context"
	"maps"
	"sync"
)

type MemoryTokenStorage struct {
	mutex sync.RWMutex
	data  map[string]string
}

func (s *MemoryTokenStorage) Get(ctx context.Context, key string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.data == nil {
		return "", nil
	}
	return s.data[key], nil
}

func (s *MemoryTokenStorage) Update(ctx context.Context, change map[string]string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.data == nil {
		s.data = map[string]string{}
	}
	maps.Copy(s.data, change)
	return nil
}

var _ ExternalTokenStorage = (*MemoryTokenStorage)(nil)

func NewMemoryTokenStorage(data map[string]string) *MemoryTokenStorage {
	if data == nil {
		data = map[string]string{}
	}
	return &MemoryTokenStorage{
		data: data,
	}
}
