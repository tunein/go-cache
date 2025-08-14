package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CacheSuite struct {
	suite.Suite
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, &CacheSuite{})
}

// TestGet ensures setting and retrieval of the requested value to the cache
func (s *CacheSuite) TestGet() {
	testCases := []struct {
		title string
		val   float32
		key   string
		exp   time.Duration
	}{
		{
			title: "Success",
			key:   "test",
			val:   0.555,
			exp:   1 * time.Second,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.title, func() {
			var (
				validate = s.Assert()
				cc       = New[string, float32](tc.exp)
			)

			cc.Set(tc.key, tc.val)
			val, err := cc.Get(tc.key)
			validate.NoError(err)
			validate.Equal(tc.val, val)

			time.Sleep(tc.exp)
			val2, err := cc.Get(tc.key)
			validate.Error(err)
			validate.Empty(val2)
		})
	}
}

// TestGetWithLoader ensures setting and retrieval of the requested value using loader func
func (s *CacheSuite) TestGetWithLoader() {
	testCases := []struct {
		title    string
		key      string
		exp      time.Duration
		loader   LoaderFunc[string, float32]
		expected float32
	}{
		{
			title: "Get value with loader func",
			key:   "test",
			exp:   1 * time.Second,
			loader: func(s string) (float32, error) {
				if s == "test" {
					return 111.89, nil
				}
				return 0, ErrNotFound
			},
			expected: 111.89,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.title, func() {
			var (
				validate = s.Assert()
				cc       = New[string, float32](tc.exp)
			)

			cc.LoaderFunc(tc.loader)

			valid, err := cc.Get(tc.key)
			validate.NoError(err)
			validate.Equal(tc.expected, valid)

			invalid, err := cc.Get("invalid-key")
			validate.Error(err)
			validate.Equal(float32(0), invalid)
		})
	}
}

// TestKeys ensures correct KEYS array returned
func (s *CacheSuite) TestKeys() {
	var (
		validate = s.Assert()
		keys     = []int{1, 10, 100, 1000, 10000}
		cc       = New[int, bool](1 * time.Nanosecond)
	)

	for _, k := range keys {
		cc.Set(k, true)
	}

	res := cc.Keys(false)
	validate.Equal(len(keys), len(res))

	for _, k := range keys {
		validate.Contains(res, k)
	}

	res = cc.Keys(true)
	validate.Equal(0, len(res))
}

// TestSetWithExpire ensures cacheItems created with passed expiration
func (s *CacheSuite) TestSetWithExpire() {
	var (
		validate = s.Assert()
		cc       = New[int, bool](1 * time.Nanosecond)
	)

	cc.SetWithExpire(1, true, 1*time.Second)
	val, err := cc.Get(1)
	validate.NoError(err)
	validate.True(val)

	time.Sleep(1 * time.Second)
	val, err = cc.Get(1)
	validate.Error(err)
	validate.False(val)
}

// TestLen ensures correct cache length is returned
func (s *CacheSuite) TestLen() {
	var (
		validate = s.Assert()
		keys     = []int{1, 10, 100, 1000, 10000}
		cc       = New[int, bool](1 * time.Nanosecond)
	)

	for _, k := range keys {
		cc.Set(k, true)
	}

	res := cc.Len(false)
	validate.Equal(len(keys), res)

	res = cc.Len(true)
	validate.Equal(0, res)
}

// TestHas ensures that cache has saved value
func (s *CacheSuite) TestHas() {
	var (
		validate = s.Assert()
		key      = 280
		cc       = New[int, bool](500 * time.Millisecond)
	)

	cc.Set(key, true)
	validate.True(cc.Has(key))
	validate.False(cc.Has(2))

	// wait until the cached item is expired
	time.Sleep(500 * time.Millisecond)
	validate.False(cc.Has(key))
}

// TestRemove ensures item is removed from the cache
func (s *CacheSuite) TestRemove() {
	var (
		validate = s.Assert()
		key      = 280
		cc       = New[int, bool](2000 * time.Millisecond)
	)

	cc.Set(key, true)
	validate.True(cc.Has(key))
	validate.Equal(1, cc.Len(false))

	cc.Remove(key)
	validate.False(cc.Has(key))
	validate.Equal(0, cc.Len(false))
}

// TestPurge ensures correct cache cleaning
func (s *CacheSuite) TestPurge() {
	var (
		validate = s.Assert()
		keys     = []int{1, 10, 100, 1000, 10000}
		cc       = New[int, bool](1 * time.Nanosecond)
	)

	for _, k := range keys {
		cc.Set(k, true)
	}

	validate.Equal(len(keys), cc.Len(false))

	cc.Purge()
	validate.Empty(cc.Len(false))
}

func (s *CacheSuite) TestSet() {
	cc := New[string, int](time.Second)

	cc.Set("a", 10)
	v, _ := cc.Get("a")
	s.Require().Equal(10, v)

	cc.Set("a", 20)
	v, _ = cc.Get("a")
	s.Require().Equal(20, v)
}

func (s *CacheSuite) TestCalcSet() {
	cc := New[string, int](time.Second)

	cc.Set("a", 5)
	cc.Update("a", func(v int) int {
		return v * v
	})

	v, _ := cc.Get("a")
	s.Require().Equal(25, v)
}

func (s *CacheSuite) TestConcurrentUpdate() {
	var (
		validate = s.Assert()
		cc       = New[int, int](1000 * time.Millisecond)
		calc     = func(in int) int {
			return in + 100
		}
		k1 = 1
	)

	cc.Set(k1, 1)
	for i := 0; i < 10; i++ {
		go func() {
			cc.Update(k1, calc)
		}()
	}
	time.Sleep(20 * time.Millisecond)
	res, err := cc.Get(k1)
	validate.NoError(err)
	validate.Equal(1001, res)
}
