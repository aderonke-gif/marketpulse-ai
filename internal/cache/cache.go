package cache

import (
	"sync"
	"time"

	"github.com/olamitunde/marketpulse-ai/internal/model"
)

type item struct {
	value     model.Analysis
	expiresAt time.Time
}

// Cache is a small TTL cache that prevents repeated inference for duplicate news.
// A production deployment can replace it with Redis without changing API handlers.
type Cache struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[string]item
}

func New(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl, items: make(map[string]item)}
}

func (c *Cache) Get(key string) (model.Analysis, bool) {
	c.mu.RLock()
	v, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(v.expiresAt) {
		if ok {
			c.mu.Lock()
			delete(c.items, key)
			c.mu.Unlock()
		}
		return model.Analysis{}, false
	}
	return v.value, true
}

func (c *Cache) Set(key string, value model.Analysis) {
	c.mu.Lock()
	c.items[key] = item{value: value, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}
