package cache

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/exp/constraints"
)

// ErrNotFound returned is nothing has been found by the provided key
var ErrNotFound = errors.New("item has not been found in the cache")

// Cache provides in-memory lru (discards the least recently used items first) cache functionality
type Cache[TKey constraints.Ordered, TValue any] struct {
	mtx              sync.RWMutex
	items            map[TKey]cacheItem[TValue]
	ttl              time.Duration
	addedFunc        AddedFunc[TKey, TValue]
	loaderExpireFunc LoaderExpireFunc[TKey, TValue]
	loaderFunc       LoaderFunc[TKey, TValue]
	loadGroup        Group[TKey, TValue]
}

// New returns reference to typed  in-memory cache instance
func New[TKey constraints.Ordered, TValue any](exp time.Duration) *Cache[TKey, TValue] {
	cc := &Cache[TKey, TValue]{}
	cc.init(exp)
	return cc
}

func (c *Cache[TKey, TValue]) init(exp time.Duration) {
	c.items = make(map[TKey]cacheItem[TValue])
	c.mtx = sync.RWMutex{}
	c.ttl = exp
	c.loadGroup = Group[TKey, TValue]{
		mtx: sync.Mutex{},
		c:   c,
	}
}

func (c *Cache[TKey, TValue]) set(key TKey, value TValue, ttl time.Duration) {
	// Check for existing item
	item, ok := c.items[key]
	if ttl == 0 {
		ttl = c.ttl
	}
	if !ok {
		item = cacheItem[TValue]{
			val:   value,
			ttl:   ttl,
			added: time.Now(),
		}
		c.items[key] = item
	}

	item.ttl = ttl

	if c.ttl > 0 && item.ttl < 1 {
		item.ttl = c.ttl
	}

	if c.addedFunc != nil {
		c.addedFunc(key, value)
	}
}

// set a new key-value pair
func (c *Cache[TKey, TValue]) Set(key TKey, value TValue) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.set(key, value, 0)
}

// Set a new key-value pair with an expiration time
func (c *Cache[TKey, TValue]) SetWithExpire(key TKey, value TValue, expiration time.Duration) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.set(key, value, expiration)
}

// Get a value from cache pool using key if it exists.
// If it does not exists key and has LoaderFunc,
// generate a value using `LoaderFunc` method returns value.
func (c *Cache[TKey, TValue]) Get(key TKey) (TValue, error) {
	v, err := c.get(key, false)
	if err == ErrNotFound {
		return c.getWithLoader(key, true)
	}
	return v, err
}

// GetIFPresent gets a value from cache pool using key if it exists.
// If it does not exists key, returns ErrNotFound.
// And send a request which refresh value for specified key if cache object has LoaderFunc.
func (c *Cache[TKey, TValue]) GetIFPresent(key TKey) (TValue, error) {
	v, err := c.get(key, false)
	if err == ErrNotFound {
		return c.getWithLoader(key, false)
	}
	return v, err
}

func (c *Cache[TKey, TValue]) get(key TKey, onLoad bool) (TValue, error) {
	v, err := c.getValue(key, onLoad)
	if err != nil {
		var def TValue
		return def, err
	}
	return v, nil
}

func (c *Cache[TKey, TValue]) getValue(key TKey, onLoad bool) (TValue, error) {
	c.mtx.Lock()
	item, ok := c.items[key]
	if ok {
		if !item.expired() {
			c.mtx.Unlock()
			return item.val, nil
		}
		delete(c.items, key)
	}
	c.mtx.Unlock()
	var def TValue
	return def, ErrNotFound
}

func (c *Cache[TKey, TValue]) getWithLoader(key TKey, isWait bool) (TValue, error) {
	var def TValue
	if c.loaderExpireFunc == nil {
		return def, ErrNotFound
	}
	value, _, err := c.load(key, func(v TValue, expiration *time.Duration, e error) (TValue, error) {
		if e != nil {
			return def, e
		}
		c.mtx.Lock()
		defer c.mtx.Unlock()
		c.set(key, v, 0)
		return v, nil
	}, isWait)
	if err != nil {
		return def, err
	}
	return value, nil
}

// load a new value using by specified key.
func (c *Cache[TKey, TValue]) load(key TKey, cb func(TValue, *time.Duration, error) (TValue, error), isWait bool) (TValue, bool, error) {
	v, called, err := c.loadGroup.Do(key, func() (v TValue, e error) {
		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("Loader panics: %v", r)
			}
		}()
		return cb(c.loaderExpireFunc(key))
	}, isWait)
	if err != nil {
		var def TValue
		return def, called, err
	}
	return v, called, nil
}

// Has checks if key exists in cache
func (c *Cache[TKey, TValue]) Has(key TKey) bool {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.has(key)
}

func (c *Cache[TKey, TValue]) has(key TKey) bool {
	item, ok := c.items[key]
	if !ok {
		return false
	}
	return !item.expired()
}

// Remove removes the provided key from the cache.
func (c *Cache[TKey, TValue]) Remove(key TKey) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	return c.remove(key)
}

func (c *Cache[TKey, TValue]) remove(key TKey) bool {
	if _, ok := c.items[key]; ok {
		delete(c.items, key)
		return true
	}
	return false
}

// Keys returns a slice of the keys in the cache.
func (c *Cache[TKey, TValue]) Keys(checkExpired bool) []TKey {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	keys := make([]TKey, 0, len(c.items))
	for k := range c.items {
		if !checkExpired || c.has(k) {
			keys = append(keys, k)
		}
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *Cache[TKey, TValue]) Len(checkExpired bool) int {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	if !checkExpired {
		return len(c.items)
	}
	var length int
	for k := range c.items {
		if c.has(k) {
			length++
		}
	}
	return length
}

// Completely clear the cache
func (c *Cache[TKey, TValue]) Purge() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.init(c.ttl)
}
