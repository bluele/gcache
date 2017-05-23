package gcache

import (
	"fmt"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
)

func buildARCache(size int) Cache {
	return New(size).
		ARC().
		EvictedFunc(evictedFuncForARC).
		Build()
}

func buildLoadingARCache(size int) Cache {
	return New(size).
		ARC().
		LoaderFunc(loader).
		EvictedFunc(evictedFuncForARC).
		Build()
}

func buildLoadingARCacheWithExpiration(clock clockwork.Clock, size int, ep time.Duration) Cache {
	return New(size).
		Clock(clock).
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
	fakeClock := clockwork.NewFakeClock()
	expTime := 5 * time.Minute
	gc := buildLoadingARCacheWithExpiration(fakeClock, 2, expTime)
	gc.Get("test1")
	gc.Get("test2")
	gc.Get("test3")
	length := gc.Len()
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", expectedLength, length)
	}

	// Advance the clock past the expiration time
	fakeClock.Advance(expTime + time.Second)
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
	testGetIFPresent(t, TYPE_ARC)
}

func TestARCGetALL(t *testing.T) {
	testGetALL(t, TYPE_ARC)
}
