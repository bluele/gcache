package gcache_test

import (
	"fmt"
	gcache "github.com/bluele/gcache"
	"testing"
	"time"
)

func buildSimpleCache(size int) gcache.Cache {
	return gcache.New(size).
		Simple().
		EvictedFunc(evictedFuncForSimple).
		Build()
}

func buildLoadingSimpleCache(size int, loader gcache.LoaderFunc) gcache.Cache {
	return gcache.New(size).
		LoaderFunc(loader).
		Simple().
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
	testGetCache(t, buildLoadingSimpleCache(size, loader), numbers)
}

func TestSimpleLength(t *testing.T) {
	gc := buildLoadingSimpleCache(1000, loader)
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
	gc := buildLoadingSimpleCache(cacheSize, loader)

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestSimpleGetIFPresent(t *testing.T) {
	cache := gcache.
		New(8).
		LoaderFunc(
		func(key interface{}) (interface{}, error) {
			time.Sleep(100 * time.Millisecond)
			return "value", nil
		}).
		Simple().
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

func TestSimpleGetALL(t *testing.T) {
	cache := gcache.
		New(8).
		Simple().
		Build()

	for i := 0; i < 8; i++ {
		cache.Set(i, i*i)
	}
	m := cache.GetALL()
	for i := 0; i < 8; i++ {
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
}
