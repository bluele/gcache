package gcache

import (
	"fmt"
	"testing"
	"time"
)

func evictedFuncForLFU(key, value interface{}) {
	fmt.Printf("[LFU] Key:%v Value:%v will be evicted.\n", key, value)
}

func buildLFUCache(size int) (Cache, error) {
	return New(size).
		LFU().
		EvictedFunc(evictedFuncForLFU).
		Expiration(time.Second).
		Build()
}

func buildLoadingLFUCache(size int, loader LoaderFunc) (Cache, error) {
	return New(size).
		LFU().
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForLFU).
		Expiration(time.Second).
		Build()
}

func TestLFUGet(t *testing.T) {
	size := 1000
	numbers := 1000

	gc, err := buildLoadingLFUCache(size, loader)
	if err != nil {
		t.Error(err)
	}

	testSetCache(t, gc, numbers)
	testGetCache(t, gc, numbers)
}

func TestLoadingLFUGet(t *testing.T) {
	size := 1000
	numbers := 1000

	gc, err := buildLoadingLFUCache(size, loader)
	if err != nil {
		t.Error(err)
	}

	testGetCache(t, gc, numbers)
}

func TestLFULength(t *testing.T) {
	gc, err := buildLoadingLFUCache(1000, loader)
	if err != nil {
		t.Error(err)
	}

	gc.Get("test1")
	gc.Get("test2")
	length := gc.Len()
	expectedLength := 2
	if gc.Len() != expectedLength {
		t.Errorf("Expected length is %v, not %v", length, expectedLength)
	}
}

func TestLFUEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc, err := buildLoadingLFUCache(cacheSize, loader)
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

func TestLFUGetIFPresent(t *testing.T) {
	testGetIFPresent(t, TYPE_LFU)
}

func TestLFUGetALL(t *testing.T) {
	testGetALL(t, TYPE_LFU)
}
