package gcache_test

import (
	"fmt"
	gcache "github.com/bluele/gcache"
	"sync/atomic"
	"testing"
	"time"
)

func buildSimpleCache(size int) gcache.Cache {
	return gcache.New(size).
		Simple().
		EvictedFunc(evictedFuncForSimple).
		Build()
}

func buildLoadingSimpleCache(size int, loader gcache.LoaderFunc, expiration time.Duration, asyncRefresh bool) gcache.Cache {
	return gcache.New(size).
		LoaderFunc(loader, asyncRefresh).
		Simple().
		Expiration(expiration).
		EvictedFunc(evictedFuncForSimple).
		Build()
}

func evictedFuncForSimple(key, value interface{}) {
	fmt.Printf("[Simple] Key:%v Value:%v will evicted.\n", key, value)
}

func TestSimpleGet(t *testing.T) {
	size := 1000
	gc := buildSimpleCache(size)
	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingSimpleGet(t *testing.T) {
	size := 1000
	numbers := 1000
	testGetCache(t, buildLoadingSimpleCache(size, loader, time.Hour, false), numbers)
}

func TestSimpleLength(t *testing.T) {
	gc := buildLoadingSimpleCache(1000, loader, time.Hour, false)
	gc.Get("test1")
	gc.Get("test2")
	length := gc.Len()
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", length, expectedLength)
	}
}

func TestSimpleEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc := buildLoadingSimpleCache(cacheSize, loader, time.Hour, false)

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestUseOldCacheWhileLoading(t *testing.T) {
	var key = "test"
	var counter int32 = 0
	var cacheExpire = 3 * time.Second
	var loadingTime = time.Second

	gc := buildLoadingSimpleCache(10, func(key interface{}) (interface{}, error) {
		time.Sleep(loadingTime - 10*time.Millisecond)
		atomic.AddInt32(&counter, 1)
		return counter, nil
	}, cacheExpire, true)
	// warmup
	gc.Get(key)

	time.Sleep(loadingTime)
	beforeCounter := counter
	// completed to load new value
	for i := 0; i < 100; i++ {
		if v, _ := gc.Get(key); v != counter {
			t.Errorf("Expected value is %v, not %v", v, counter)
		}
	}
	time.Sleep(cacheExpire - loadingTime)

	// before load new value.
	for i := 0; i < 10; i++ {
		if v, _ := gc.Get(key); v != beforeCounter {
			t.Errorf("%v != %v", v, beforeCounter)
		}
	}
	time.Sleep(loadingTime)

	// completed to load new value
	if v, _ := gc.Get(key); v != counter {
		t.Errorf("Expected value is %v, not %v", v, counter)
	}
}
