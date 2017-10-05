package gcache

import (
	"fmt"
	"testing"
	"time"
)

func evictedFuncForLRU(key, value interface{}) {
	fmt.Printf("[LRU] Key:%v Value:%v will be evicted.\n", key, value)
}

func buildLRUCache(size int) Cache {
	return New(size).
		LRU().
		EvictedFunc(evictedFuncForLRU).
		Expiration(time.Second).
		Build()
}

func buildLoadingLRUCache(size int, loader LoaderFunc) Cache {
	return New(size).
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
	testGetIFPresent(t, TYPE_LRU)
}

func TestLRUGetALL(t *testing.T) {
	testGetALL(t, TYPE_LRU)
}
