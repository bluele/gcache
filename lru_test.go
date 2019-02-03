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
		Build()
}

func buildLoadingLRUCache(size int, loader LoaderFunc) Cache {
	return New(size).
		LRU().
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForLRU).
		Build()
}

func buildLoadingLRUCacheWithExpiration(size int, ep time.Duration) Cache {
	return New(size).
		LRU().
		Expiration(ep).
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForLRU).
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

func TestLRUHas(t *testing.T) {
	gc := buildLoadingLRUCacheWithExpiration(2, time.Millisecond)

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			gc.Get("test1")
			gc.Get("test2")

			if gc.Has("test0") {
				t.Fatal("should not have test0")
			}
			if !gc.Has("test1") {
				t.Fatal("should have test1")
			}
			if !gc.Has("test2") {
				t.Fatal("should have test2")
			}

			time.Sleep(time.Millisecond)

			if gc.Has("test0") {
				t.Fatal("should not have test0")
			}
			if gc.Has("test1") {
				t.Fatal("should not have test1")
			}
			if gc.Has("test2") {
				t.Fatal("should not have test2")
			}
		})
	}
}
