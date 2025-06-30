package redis

import "sync"

type KVStore struct {
	mutex sync.RWMutex
	store map[string]string
}

func NewStore() *KVStore {
	return &KVStore{
		store: make(map[string]string),
	}
}

func (kvStore *KVStore) Set(key string, value string) string {
	kvStore.mutex.Lock()
	defer kvStore.mutex.Unlock()
	kvStore.store[key] = value
	return ""
}

func (kvStore *KVStore) Get(key string) string {
	kvStore.mutex.RLock()
	defer kvStore.mutex.RUnlock()
	value, ok := kvStore.store[key]
	if ok {
		return value
	}
	return ""
}
