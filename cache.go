package main

import (
	"sync"
	"time"
)

// entry - typical element of cache
type entry struct {
	filename string
	filepath string
	expiry   *time.Time
}

// Cache - simple implementation of cache
// More information: https://en.wikipedia.org/wiki/Time_to_live
type Cache struct {
	timeTTL              time.Duration
	cache                map[string]*entry
	lock                 *sync.RWMutex
	BeforeDeleteCallback func(*entry)
}

// NewCache - initialization of new cache.
// For avoid mistake - minimal time to live is 1 minute.
// For simplification, - key is string and cache haven`t stop method
func NewCache(interval time.Duration) *Cache {
	if interval < time.Second {
		interval = time.Second
	}
	cache := &Cache{
		timeTTL: interval,
		cache:   make(map[string]*entry),
		lock:    &sync.RWMutex{},
	}
	go func() {
		ticker := time.NewTicker(cache.timeTTL)
		for {
			// wait of ticker
			now := <-ticker.C

			// remove entry outside TTL
			cache.lock.Lock()
			for id, entry := range cache.cache {
				if entry.expiry != nil && entry.expiry.Before(now) {
					cache.BeforeDeleteCallback(entry)
					delete(cache.cache, id)
				}
			}
			cache.lock.Unlock()
		}
	}()
	return cache
}

// Count - return amount element of TTL map.
func (cache *Cache) Count() int {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	return len(cache.cache)
}

// Get - return value from cache
func (cache *Cache) Get(key string) *entry {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	e, ok := cache.cache[key]

	if ok && e.expiry != nil && e.expiry.After(time.Now()) {
		return e
	}
	return nil
}

// Add - add key/value in cache
func (cache *Cache) Add(key string, filepath, filename string, ttl time.Duration) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	expiry := time.Now().Add(ttl)

	cache.cache[key] = &entry{
		filepath: filepath,
		filename: filename,
		expiry:   &expiry,
	}
}

// GetKeys - return all keys of cache map
func (cache *Cache) GetKeys() []interface{} {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	keys := make([]interface{}, len(cache.cache))
	var i int
	for k := range cache.cache {
		keys[i] = k
		i++
	}
	return keys
}
