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
	entries  map[string]cacheEntry
	mu       *sync.Mutex
	interval time.Duration
}

func NewCache(i time.Duration) Cache {
	c := Cache{
		entries:  map[string]cacheEntry{},
		mu:       &sync.Mutex{},
		interval: i,
	}
	c.reapLoop()
	return c
}

func (c Cache) Add(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       value,
	}
}

func (c Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.entries[key]
	if ok {
		return v.val, true
	} else {
		return nil, false
	}
}

func (c Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	go func() {
		time.Sleep(c.interval)
		for range ticker.C {
			expiry := time.Now().Add(-c.interval)
			c.mu.Lock()
			for k, v := range c.entries {
				if v.createdAt.Compare(expiry) < 0 {
					delete(c.entries, k)
				}
			}
			c.mu.Unlock()
		}
	}()

	ticker.Reset(c.interval)
}
