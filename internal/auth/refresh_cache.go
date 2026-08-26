package auth

import (
	"sync"
	"time"
)

// refreshResult is what we hand back to a request that hit the cache
// instead of the database.
type refreshResult struct {
	Access  string
	Refresh string
}

// refreshCache absorbs the case where several requests race in with the
// same (now-rotated) refresh token at nearly the same instant. Without it,
// the second request would see the token already replaced in the DB and be
// wrongly treated as reuse. It is NOT a substitute for the DB-level lock in
// GetForUpdate — it just lets the "loser" of the race return the same
// result as the "winner" within a short window.
//
// This is in-process only. If you run multiple API instances behind a load
// balancer, back this with Redis instead so all instances share it.
type refreshCache struct {
	mu   sync.Mutex
	data map[string]cacheEntry
	ttl  time.Duration
}

type cacheEntry struct {
	result  refreshResult
	expires time.Time
}

func newRefreshCache(ttl time.Duration) *refreshCache {
	c := &refreshCache{data: make(map[string]cacheEntry), ttl: ttl}
	go c.sweep()
	return c
}

func (c *refreshCache) Get(key string) (refreshResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.data[key]
	if !ok || time.Now().After(e.expires) {
		return refreshResult{}, false
	}
	return e.result, true
}

func (c *refreshCache) Set(key string, result refreshResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = cacheEntry{result: result, expires: time.Now().Add(c.ttl)}
}

func (c *refreshCache) sweep() {
	t := time.NewTicker(time.Minute)
	for range t.C {
		c.mu.Lock()
		now := time.Now()
		for k, e := range c.data {
			if now.After(e.expires) {
				delete(c.data, k)
			}
		}
		c.mu.Unlock()
	}
}
