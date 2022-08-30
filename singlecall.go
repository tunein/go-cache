package cache

// This module provides a duplicate function call suppression
// mechanism.
// original idea taken from here:
// https://github.com/bluele/gcache/blob/master/singleflight.go

import (
	"sync"

	"golang.org/x/exp/constraints"
)

// call is an in-flight or completed Do call
type call[TValue any] struct {
	wg  sync.WaitGroup
	val TValue
	err error
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type Group[TKey constraints.Ordered, TValue any] struct {
	c   ICache[TKey, TValue]
	mtx sync.Mutex             // protects m
	m   map[TKey]*call[TValue] // lazily initialized
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *Group[TKey, TValue]) Do(key TKey, fn func() (TValue, error), isWait bool) (TValue, bool, error) {
	var def TValue
	g.mtx.Lock()
	v, err := g.c.get(key, true)
	if err == nil {
		g.mtx.Unlock()
		return v, false, nil
	}
	if g.m == nil {
		g.m = make(map[TKey]*call[TValue])
	}
	if c, ok := g.m[key]; ok {
		g.mtx.Unlock()
		if !isWait {
			return def, false, ErrNotFound
		}
		c.wg.Wait()
		return c.val, false, c.err
	}
	c := new(call[TValue])
	c.wg.Add(1)
	g.m[key] = c
	g.mtx.Unlock()
	if !isWait {
		go g.call(c, key, fn)
		return def, false, ErrNotFound
	}
	v, err = g.call(c, key, fn)
	return v, true, err
}

func (g *Group[TKey, TValue]) call(c *call[TValue], key TKey, fn func() (TValue, error)) (TValue, error) {
	c.val, c.err = fn()
	c.wg.Done()

	g.mtx.Lock()
	delete(g.m, key)
	g.mtx.Unlock()

	return c.val, c.err
}
