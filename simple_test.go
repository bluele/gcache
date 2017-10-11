package gcache

import (
	"fmt"
	"testing"
)

func buildSimpleCache(size int) (Cache, error) {
	return New(size).
		Simple().
		EvictedFunc(evictedFuncForSimple).
		Build()
}

func buildLoadingSimpleCache(size int, loader LoaderFunc) (Cache, error) {
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
	gc, err := buildSimpleCache(size)
	if err != nil {
		t.Error(err)
	}

	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingSimpleGet(t *testing.T) {
	size := 1000
	numbers := 1000
	gc, err := buildLoadingSimpleCache(size, loader)
	if err != nil {
		t.Error(err)
	}

	testGetCache(t, gc, numbers)
}

func TestSimpleLength(t *testing.T) {
	gc, err := buildLoadingSimpleCache(1000, loader)
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

func TestSimpleEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc, err := buildLoadingSimpleCache(cacheSize, loader)
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

func TestSimpleUnboundedNoEviction(t *testing.T) {
	numbers := 1000
	size_tracker := 0
	gcu, err := buildLoadingSimpleCache(0, loader)
	if err != nil {
		t.Error(err)
	}

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
