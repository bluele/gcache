package gcache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/nethack42/gcache"
)

func evictedFuncForLRU(key, value interface{}) {
	fmt.Printf("[LRU] Key:%v Value:%v will evicted.\n", key, value)
}

func buildLRUCache(size int) gcache.Cache {
	return gcache.New(size).
		LRU().
		EvictedFunc(evictedFuncForLRU).
		Expiration(time.Second).
		Build()
}

func buildLoadingLRUCache(size int, loader gcache.LoaderFunc) gcache.Cache {
	return gcache.New(size).
		LRU().
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForLRU).
		Expiration(time.Second).
		Build()
}

func TestLRUGet(t *testing.T) {
	size := 1000
	gc := buildLRUCache(size)
	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingLRUGet(t *testing.T) {
	size := 1000
	gc := buildLoadingLRUCache(size, loader)
	testGetCache(t, gc, size)
}

func TestLRULength(t *testing.T) {
	gc := buildLoadingLRUCache(1000, loader)
	gc.Get("test1")
	gc.Get("test2")
	length := gc.Len()
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", length, expectedLength)
	}
}

func TestLRUEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc := buildLoadingLRUCache(cacheSize, loader)

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestLRUGetIFPresent(t *testing.T) {
	cache := gcache.
		New(8).
		LoaderFunc(
		func(key interface{}) (interface{}, error) {
			time.Sleep(100 * time.Millisecond)
			return "value", nil
		}).
		LRU().
		Build()

	v, err := cache.GetIFPresent("key")
	if err != gcache.KeyNotFoundError {
		t.Errorf("err should not be %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	v, err = cache.GetIFPresent("key")
	if err != nil {
		t.Errorf("err should not be %v", err)
	}
	if v != "value" {
		t.Errorf("v should not be %v", v)
	}
}

func TestLRUGetALL(t *testing.T) {
	size := 8
	cache := gcache.
		New(size).
		LRU().
		Build()

	for i := 0; i < size; i++ {
		cache.Set(i, i*i)
	}
	m := cache.GetALL()
	for i := 0; i < size; i++ {
		v, ok := m[i]
		if !ok {
			t.Errorf("m should contain %v", i)
			continue
		}
		if v.(int) != i*i {
			t.Errorf("%v != %v", v, i*i)
			continue
		}
	}

	cache.Stop()
}
