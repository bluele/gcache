package gcache

import (
	"fmt"
	"testing"
	"time"
)

func buildARCache(size int) (Cache, error) {
	return New(size).
		ARC().
		EvictedFunc(evictedFuncForARC).
		Build()
}

func buildLoadingARCache(size int) (Cache, error) {
	return New(size).
		ARC().
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForARC).
		Build()
}

func buildLoadingARCacheWithExpiration(size int, ep time.Duration) (Cache, error) {
	return New(size).
		ARC().
		Expiration(ep).
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForARC).
		Build()
}

func evictedFuncForARC(key, value interface{}) {
	fmt.Printf("[ARC] Key:%v Value:%v will be evicted.\n", key, value)
}

func TestARCGet(t *testing.T) {
	size := 1000
	gc, err := buildARCache(size)
	if err != nil {
		t.Error(err)
	}
	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingARCGet(t *testing.T) {
	size := 1000
	numbers := 1000
	gc, err := buildLoadingARCache(size)
	if err != nil {
		t.Error(err)
	}
	testGetCache(t, gc, numbers)
}

func TestARCLength(t *testing.T) {
	gc, err := buildLoadingARCacheWithExpiration(2, time.Millisecond)
	if err != nil {
		t.Error(err)
	}

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
	numbers := cacheSize + 1
	gc, err := buildLoadingARCache(cacheSize)
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

func TestARCPurgeCache(t *testing.T) {
	cacheSize := 10
	purgeCount := 0
	gc, err := New(cacheSize).
		ARC().
		LoaderFunc(loader).
		PurgeVisitorFunc(func(k, v interface{}) {
			purgeCount++
		}).
		Build()

	if err != nil {
		t.Error(err)
	}

	for i := 0; i < cacheSize; i++ {
		_, err := gc.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	gc.Purge()

	if purgeCount != cacheSize {
		t.Errorf("failed to purge everything")
	}
}

func TestARCGetIFPresent(t *testing.T) {
	testGetIFPresent(t, TYPE_ARC)
}

func TestARCGetALL(t *testing.T) {
	testGetALL(t, TYPE_ARC)
}
