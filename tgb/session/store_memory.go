package session

import (
	"context"
	"sync"
)

// StoreMemory is a memory storage for sessions.
// It implements [Store] and is thread-safe.
type StoreMemory struct {
	kv   map[string][]byte
	lock sync.Mutex
}

var _ Store = (*StoreMemory)(nil)

func NewStoreMemory() *StoreMemory {
	return &StoreMemory{
		kv: make(map[string][]byte),
	}
}

func (s *StoreMemory) Set(ctx context.Context, key string, value []byte) error {
	s.lock.Lock()
	s.kv[key] = value
	s.lock.Unlock()

	return nil
}

func (s *StoreMemory) Get(ctx context.Context, key string) ([]byte, error) {
	s.lock.Lock()
	v, ok := s.kv[key]
	s.lock.Unlock()

	if !ok {
		return nil, nil
	}

	return v, nil
}

func (s *StoreMemory) Del(ctx context.Context, key string) error {
	s.lock.Lock()
	delete(s.kv, key)
	s.lock.Unlock()

	return nil
}
