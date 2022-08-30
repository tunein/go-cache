package cache

import "time"

type cacheItem[V any] struct {
	val   V
	ttl   time.Duration
	added time.Time
}

func (c cacheItem[V]) expired() bool {
	if c.ttl <= 0 {
		return false
	}

	return time.Since(c.added) > c.ttl
}
