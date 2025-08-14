package cache

import (
	"time"
)

type (
	// AddedFunc function that should be called when new item is added to the cache
	AddedFunc[TKey comparable, TValue any] func(TKey, TValue)
	// LoaderFunc function that is used to load missing or expired cache entry
	LoaderFunc[TKey comparable, TValue any] func(TKey) (TValue, error)
	// LoaderExpireFunc is called when item is expired
	LoaderExpireFunc[TKey comparable, TValue any] func(TKey) (TValue, *time.Duration, error)
)

// LoaderFunc: create a new value with this function if cached value is expired.
func (c *Cache[TKey, TValue]) LoaderFunc(loaderFunc LoaderFunc[TKey, TValue]) *Cache[TKey, TValue] {
	c.loaderExpireFunc = func(k TKey) (TValue, *time.Duration, error) {
		v, err := loaderFunc(k)
		return v, nil, err
	}
	return c
}

// LoaderExpireFunc - loader function with expiration, create a new value with this function if cached value is expired.
// If nil returned instead of time.Duration from loaderExpireFunc than value will never expire.
func (c *Cache[TKey, TValue]) LoaderExpireFunc(loaderExpireFunc LoaderExpireFunc[TKey, TValue]) *Cache[TKey, TValue] {
	c.loaderExpireFunc = loaderExpireFunc
	return c
}

// AddedFunc - if provided, this function will be called after each new value is added to the cache
func (c *Cache[TKey, TValue]) AddedFunc(addedFunc AddedFunc[TKey, TValue]) *Cache[TKey, TValue] {
	c.addedFunc = addedFunc
	return c
}
