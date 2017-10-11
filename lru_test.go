package gcache

import (
	"fmt"
	"testing"
	"time"
)

func evictedFuncForLRU(key, value interface{}) {
	fmt.Printf("[LRU] Key:%v Value:%v will be evicted.\n", key, value)
}

func buildLRUCache(size int) (Cache, error) {
	return New(size).
		LRU().
		EvictedFunc(evictedFuncForLRU).
		Expiration(time.Second).
		Build()
}

func buildLoadingLRUCache(size int, loader LoaderFunc) (Cache, error) {
	return New(size).
		LRU().
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForLRU).
		Expiration(time.Second).
		Build()
}

func TestLRUGet(t *testing.T) {
	size := 1000
	gc, err := buildLRUCache(size)
	if err != nil {
		t.Error(err)
	}

	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingLRUGet(t *testing.T) {
	size := 1000
	gc, err := buildLoadingLRUCache(size, loader)
	if err != nil {
		t.Error(err)
	}

	testGetCache(t, gc, size)
}

func TestLRULength(t *testing.T) {
	gc, err := buildLoadingLRUCache(1000, loader)
	if err != nil {
		t.Error(err)
	}
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
	gc, err := buildLoadingLRUCache(cacheSize, loader)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < numbers; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestLRUGetIFPresent(t *testing.T) {
	testGetIFPresent(t, TYPE_LRU)
}

func TestLRUGetALL(t *testing.T) {
	testGetALL(t, TYPE_LRU)
}
