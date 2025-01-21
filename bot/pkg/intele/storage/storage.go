package storage

import "sync"

// StateStorage interface for storing user states
type StateStorage interface {
	Set(userID int64, state string) error
	Get(userID int64) (string, error)
	Delete(userID int64)
}

// MemoryStorage is a simple in-memory implementation of StateStorage
type MemoryStorage struct {
	mu    sync.RWMutex
	store map[int64]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		store: make(map[int64]string),
	}
}

func (m *MemoryStorage) Set(userID int64, state string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[userID] = state
	return nil
}

func (m *MemoryStorage) Get(userID int64) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, ok := m.store[userID]
	if !ok {
		return "", nil
	}
	return state, nil
}

func (m *MemoryStorage) Delete(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, userID)
}
