package redis

import "sync"

type KVStore struct {
	mutex sync.Mutex
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
  kvStore.mutex.Lock()
  defer kvStore.mutex.Unlock()
	value, ok := kvStore.store[key]
	if ok {
		return value
	}
	return ""
}
