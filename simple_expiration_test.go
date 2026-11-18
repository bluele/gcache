package gcache_test

import (
	"sync"
	"testing"
	"time"

	"github.com/puper/gcache"
)

func TestCacheReclaimsExpiredTailWithoutMoreTraffic(t *testing.T) {
	for _, build := range []func() gcache.Cache{
		func() gcache.Cache {
			c := gcache.New(0).Build()
			c.SetWithExpire("one", make([]byte, 1<<20), time.Millisecond)
			return c
		},
		func() gcache.Cache {
			c := gcache.New(0).Expiration(time.Millisecond).Build()
			c.Set("one", 1)
			return c
		},
		func() gcache.Cache {
			c := gcache.New(0).LoaderExpireFunc(func(interface{}) (interface{}, *time.Duration, error) { d := time.Millisecond; return 1, &d, nil }).Build()
			c.Get("one")
			return c
		},
	} {
		c := build()
		deadline := time.Now().Add(2 * time.Second)
		for c.Len(false) > 0 && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
		}
		if c.Len(false) != 0 {
			t.Fatal("expired tail still retains its value without further cache operations")
		}
	}
}

func TestCacheExpiryConcurrentWritesAndPurge(t *testing.T) {
	c := gcache.New(0).Build()
	done := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		for worker := 0; worker < 4; worker++ {
			wg.Add(1)
			go func(worker int) {
				defer wg.Done()
				for i := 0; i < 12000; i++ {
					c.SetWithExpire(worker*12000+i, i, -6*time.Second)
				}
			}(worker)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10000; i++ {
				c.SetWithExpire("refresh", i, time.Hour)
				c.Get("refresh")
				if i%100 == 0 {
					c.Purge()
				}
			}
		}()
		wg.Wait()
		c.SetWithExpire("keep", 1, time.Hour)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(8 * time.Second):
		t.Fatal("cache deadlocked during expiry/write/purge")
	}
	time.Sleep(1100 * time.Millisecond)
	if v, err := c.Get("keep"); err != nil || v != 1 {
		t.Fatalf("fresh value lost: %v %v", v, err)
	}
	c.Purge()
}

func TestSizedSimpleCacheBoundsLiveValues(t *testing.T) {
	c := gcache.New(8).Build()
	for i := 0; i < 100; i++ {
		c.SetWithExpire(i, i, time.Hour)
	}
	if c.Len(false) > 8 {
		t.Fatalf("capacity exceeded: %d", c.Len(false))
	}
	c.Purge()
}

// More expirations than all of the former timewheel/GC queues combined, while
// writes continue after the expiry tick. Completion is the regression assertion.
func TestCacheExpiryDoesNotBlockOngoingWrites(t *testing.T) {
	c := gcache.New(0).Build()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 20000; i++ {
			c.SetWithExpire(i, 1, -6*time.Second)
		}
		until := time.Now().Add(2 * time.Second)
		for i := 0; time.Now().Before(until); i++ {
			c.SetWithExpire(30000+i%2000, 1, time.Hour)
		}
		c.Get(30000)
		c.Purge()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(6 * time.Second):
		t.Fatal("expiry queues deadlocked ongoing writes")
	}
}
