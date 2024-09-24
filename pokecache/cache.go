package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
	ttl     time.Duration
  ticker  *time.Ticker
}

func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
		ticker:  time.NewTicker(ttl),
	}

	go cache.reapLoop()

	return cache
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Since(entry.createdAt) > c.ttl {
		delete(c.entries, key)
		return nil, false
	}

	return entry.val, true
}

func (c *Cache) reapLoop() {
	for range c.ticker.C {
		c.mu.Lock()

		for key, entry := range c.entries {
			if time.Since(entry.createdAt) > c.ttl {
				delete(c.entries, key)
			}
		}

		c.mu.Unlock() 	}
}

func (c *Cache) Stop() {
	c.ticker.Stop()
}



