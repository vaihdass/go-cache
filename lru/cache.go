// Package lru implements an in-memory LRU (Least Recently Used) cache with TTL-based invalidation.
//
// TTL was chosen for cache invalidation as it allows automatic expiration of entries after
// a specified time, ensuring consistency for time-sensitive data and
// reducing the memory and CPU overhead of renewing entries
// (instead of constantly running a separate worker to check and update expired values,
// even when it may not be necessary).
package lru

import (
	"container/list"
	"sync"
	"time"
)

const defaultSize int = 128

// Cache is a generic, thread-safe cache implementing LRU eviction and TTL-based invalidation
type Cache[K comparable, V any] struct {
	items     map[K]*list.Element
	evictList *list.List
	capacity  int
	lock      sync.Mutex

	// ttl defines the time-to-live duration for cache entries, zero value means TTL is not used
	ttl time.Duration
}

func New[K comparable, V any](opts ...Option) (*Cache[K, V], error) {
	var o cacheOptions
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&o)
	}
	if o.capacity <= 0 {
		o.capacity = defaultSize
	}

	return &Cache[K, V]{
		items:     make(map[K]*list.Element),
		evictList: list.New(),
		capacity:  o.capacity,
		ttl:       o.ttl,
	}, nil
}

// Set sets a value for specified key to the cache
func (c *Cache[K, V]) Set(k K, v V) {
	expires := time.Now().Add(c.ttl)

	c.lock.Lock()
	defer c.lock.Unlock()

	e, ok := c.items[k]
	if ok {
		e.Value = cached[K, V]{
			key:       k,
			value:     v,
			expiredAt: expires,
		}
		c.evictList.MoveToFront(e)
		return
	}

	if c.evictList.Len() >= c.capacity {
		if last := c.evictList.Back(); last != nil {
			c.evictList.Remove(last)

			val := last.Value.(cached[K, V])
			delete(c.items, val.key)
		}
	}

	val := cached[K, V]{
		key:       k,
		value:     v,
		expiredAt: expires,
	}
	e = c.evictList.PushFront(val)
	c.items[k] = e
}

// Get looks up a key's value from the cache, presented = false if value expired or wasn't provided
func (c *Cache[K, V]) Get(k K) (value V, presented bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	e, ok := c.items[k]
	if !ok {
		return
	}
	val := e.Value.(cached[K, V])

	// ttl zero value means TTL is not used
	if c.ttl != 0 && val.expired(time.Now()) {
		return
	}

	c.evictList.MoveToFront(e)
	return val.value, true
}
