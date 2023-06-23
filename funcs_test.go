package cache

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CacheFuncsSuite struct {
	suite.Suite
}

func TestCacheFuncsSuite(t *testing.T) {
	suite.Run(t, &CacheFuncsSuite{})
}

// TestLoaderExpireFunc ensures loader expiration callback called correctly
func (s *CacheFuncsSuite) TestLoaderExpireFunc() {
	var (
		called   atomic.Bool
		validate = s.Assert()
		fu       = func(key int) (bool, *time.Duration, error) {
			called.Store(true)
			// simulate fetching delay
			time.Sleep(250 * time.Millisecond)
			if key == 100 {
				return true, nil, nil
			}
			return false, nil, errors.New("some error")
		}
		cc = New[int, bool](1 * time.Nanosecond)
	)
	cc.Set(100, true)
	validate.False(called.Load())
	cc.LoaderExpireFunc(fu)

	// simulating cache call from other goroutine
	go func() { _, _ = cc.Get(100) }()
	res, err := cc.Get(100)
	validate.NoError(err)
	validate.True(called.Load())
	validate.True(res)
}
