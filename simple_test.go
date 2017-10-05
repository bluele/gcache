package gcache

import (
	"fmt"
	"testing"
)

func buildSimpleCache(size int) Cache {
	return New(size).
		Simple().
		EvictedFunc(evictedFuncForSimple).
		Build()
}

func buildLoadingSimpleCache(size int, loader LoaderFunc) Cache {
	return New(size).
		LoaderFunc(loader).
		Simple().
		EvictedFunc(evictedFuncForSimple).
		Build()
}

func evictedFuncForSimple(key, value interface{}) {
	fmt.Printf("[Simple] Key:%v Value:%v will be evicted.\n", key, value)
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

func TestSimpleUnboundedNoEviction(t *testing.T) {
	numbers := 1000
	size_tracker := 0
	gcu := buildLoadingSimpleCache(0, loader)

	for i := 0; i < numbers; i++ {
		current_size := gcu.Len()
		if current_size != size_tracker {
			t.Errorf("Excepted cache size is %v not %v", current_size, size_tracker)
		}

		_, err := gcu.Get(fmt.Sprintf("Key-%d", i))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		size_tracker++
	}
}

func TestSimpleGetIFPresent(t *testing.T) {
	testGetIFPresent(t, TYPE_SIMPLE)
}

func TestSimpleGetALL(t *testing.T) {
	testGetALL(t, TYPE_SIMPLE)
}
