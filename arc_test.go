package gcache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/bluele/gcache"
)

func buildARCache(size int) gcache.Cache {
	return gcache.New(size).
		ARC().
		EvictedFunc(evictedFuncForARC).
		Build()
}

func buildLoadingARCache(size int) gcache.Cache {
	return gcache.New(size).
		ARC().
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForARC).
		Build()
}

func buildLoadingARCacheWithExpiration(size int, ep time.Duration) gcache.Cache {
	return gcache.New(size).
		ARC().
		Expiration(ep).
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForARC).
		Build()
}

func evictedFuncForARC(key, value interface{}) {
	fmt.Printf("[ARC] Key:%v Value:%v will evicted.\n", key, value)
}

func TestARCGet(t *testing.T) {
	size := 1000
	gc := buildARCache(size)
	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingARCGet(t *testing.T) {
	size := 1000
	numbers := 1000
	testGetCache(t, buildLoadingARCache(size), numbers)
}

func TestARCLength(t *testing.T) {
	gc := buildLoadingARCacheWithExpiration(2, time.Millisecond)
	gc.Get("test1")
	gc.Get("test2")
	gc.Get("test3")
	length := gc.Len()
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", expectedLength, length)
	}
	time.Sleep(time.Millisecond)
	gc.Get("test4")
	length = gc.Len()
	expectedLength = 1
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", expectedLength, length)
	}
}

func TestARCEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc := buildLoadingARCache(cacheSize)

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestARCGetIFPresent(t *testing.T) {
	cache := gcache.
		New(8).
		LoaderFunc(
			func(key interface{}) (interface{}, error) {
				time.Sleep(time.Millisecond)
				return "value", nil
			}).
		ARC().
		Build()

	v, err := cache.GetIFPresent("key")
	if err != gcache.KeyNotFoundError {
		t.Errorf("err should not be %v", err)
	}

	time.Sleep(2 * time.Millisecond)

	v, err = cache.GetIFPresent("key")
	if err != nil {
		t.Errorf("err should not be %v", err)
	}
	if v != "value" {
		t.Errorf("v should not be %v", v)
	}
}

func TestARCGetALL(t *testing.T) {
	size := 8
	cache := gcache.
		New(size).
		Expiration(time.Millisecond).
		ARC().
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
	time.Sleep(time.Millisecond)

	cache.Set(size, size*size)
	m = cache.GetALL()
	if len(m) != 1 {
		t.Errorf("%v != %v", len(m), 1)
	}
	if _, ok := m[size]; !ok {
		t.Errorf("%v should contains key '%v'", m, size)
	}
}
