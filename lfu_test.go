package gcache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/nethack42/gcache"
)

func evictedFuncForLFU(key, value interface{}) {
	fmt.Printf("[LFU] Key:%v Value:%v will evicted.\n", key, value)
}

func buildLFUCache(size int) gcache.Cache {
	return gcache.New(size).
		LFU().
		EvictedFunc(evictedFuncForLFU).
		Expiration(time.Second).
		EnableGC(time.Second).
		Build()
}

func buildLoadingLFUCache(size int, loader gcache.LoaderFunc) gcache.Cache {
	return gcache.New(size).
		LFU().
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForLFU).
		Expiration(time.Second).
		EnableGC(time.Second).
		Build()
}

func TestLFUGet(t *testing.T) {
	size := 1000
	numbers := 1000

	gc := buildLoadingLFUCache(size, loader)
	testSetCache(t, gc, numbers)
	testGetCache(t, gc, numbers)
}

func TestLoadingLFUGet(t *testing.T) {
	size := 1000
	numbers := 1000

	gc := buildLoadingLFUCache(size, loader)
	testGetCache(t, gc, numbers)
	gc.Stop()
}

func TestLFULength(t *testing.T) {
	gc := buildLoadingLFUCache(1000, loader)
	gc.Get("test1")
	gc.Get("test2")
	length := gc.Len()
	expectedLength := 2
	if gc.Len() != expectedLength {
		t.Errorf("Expected length is %v, not %v", length, expectedLength)
	}
	gc.Stop()
}

func TestLFUEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc := buildLoadingLFUCache(cacheSize, loader)

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
	gc.Stop()
}

func TestLFUGetIFPresent(t *testing.T) {
	cache := gcache.
		New(8).
		LoaderFunc(
		func(key interface{}) (interface{}, error) {
			time.Sleep(100 * time.Millisecond)
			return "value", nil
		}).
		LFU().
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

	cache.Stop()
}

func TestLFUGetALL(t *testing.T) {
	size := 8
	cache := gcache.
		New(size).
		LFU().
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
