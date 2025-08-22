// Package cache provides in-memory lru (discards the least recently used items first) cache functionality
//
//	Copyright 2025 TuneIn, Inc. All rights reserved.
//
// Use of this source code is governed by Apache License 2.0
// license that can be found in the LICENSE file.
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
