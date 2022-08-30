package cache

import (
	"time"

	"golang.org/x/exp/constraints"
)

// ICache represents cache entity
type ICache[TKey constraints.Ordered, TValue any] interface {
	// Set inserts or updates the specified key-value pair.
	Set(key TKey, value TValue)
	// SetWithExpire inserts or updates the specified key-value pair with an expiration time.
	SetWithExpire(key TKey, value TValue, expiration time.Duration)
	// Get returns the value for the specified key if it is present in the cache.
	// If the key is not present in the cache and the cache has LoaderFunc,
	// invoke the `LoaderFunc` function and inserts the key-value pair in the cache.
	// If the key is not present in the cache and the cache does not have a LoaderFunc,
	// return KeyNotFoundError.
	Get(key TKey) (TValue, error)
	// GetIFPresent returns the value for the specified key if it is present in the cache.
	// Return KeyNotFoundError if the key is not present.
	GetIFPresent(key TKey) (TValue, error)
	get(key TKey, onLoad bool) (TValue, error)
	// Remove removes the specified key from the cache if the key is present.
	// Returns true if the key was present and the key has been deleted.
	Remove(key TKey) bool
	// Purge removes all key-value pairs from the cache.
	Purge()
	// Keys returns a slice containing all keys in the cache.
	Keys(checkExpired bool) []TKey
	// Len returns the number of items in the cache.
	Len(checkExpired bool) int
	// Has returns true if the key exists in the cache.
	Has(key TKey) bool
}
