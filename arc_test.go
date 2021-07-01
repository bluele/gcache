package gcache

import (
	"fmt"
	"testing"
	"time"
)

func TestARCGet(t *testing.T) {
	size := 1000
	gc := buildTestCache(t, TYPE_ARC, size)
	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingARCGet(t *testing.T) {
	size := 1000
	numbers := 1000
	testGetCache(t, buildTestLoadingCache(t, TYPE_ARC, size, loader), numbers)
}

func TestARCLength(t *testing.T) {
	gc := buildTestLoadingCacheWithExpiration(t, TYPE_ARC, 2, time.Millisecond)
	gc.Get("test1")
	gc.Get("test2")
	gc.Get("test3")
	length := gc.Len(true)
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", expectedLength, length)
	}
	time.Sleep(time.Millisecond)
	gc.Get("test4")
	length = gc.Len(true)
	expectedLength = 1
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", expectedLength, length)
	}
}

func TestARCEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := cacheSize + 1
	gc := buildTestLoadingCache(t, TYPE_ARC, cacheSize, loader)

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
	gc := New(cacheSize).
		ARC().
		LoaderFunc(loader).
		PurgeVisitorFunc(func(k, v interface{}) {
			purgeCount++
		}).
		Build()

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

func TestARCHas(t *testing.T) {
	gc := buildTestLoadingCacheWithExpiration(t, TYPE_ARC, 2, 10*time.Millisecond)

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

			time.Sleep(20 * time.Millisecond)

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

func TestARCSizer(t *testing.T) {
	var evicts int
	evict := func(k, v interface{}) {
		evicts++
	}
	c := New(3).ARC().EvictedFunc(evict).Build()

	c.Set(1, sizerInt(1))
	c.Set(2, sizerInt(2))

	v, _ := c.Get(2)
	if v != sizerInt(2) {
		t.Fatal(v)
	}

	if evicts != 0 {
		t.Fatal(evicts)
	}
	if l := c.Len(false); l != 3 {
		t.Fatal(l)
	}

	c.Set(3, sizerInt(3))

	if evicts != 1 {
		t.Fatal(evicts)
	}
	if l := c.Len(false); l != 5 {
		t.Fatal(l)
	}

	c.Set(4, sizerInt(4))

	if evicts != 2 {
		t.Fatal(evicts)
	}
	if l := c.Len(false); l != 6 {
		t.Fatal(l)
	}

	c.Set(6, sizerInt(6))

	if evicts != 3 {
		t.Fatal(evicts)
	}
	if l := c.Len(false); l != 8 {
		t.Fatal(l)
	}

	v, _ = c.Get(6)
	if v != sizerInt(6) {
		t.Fatal(v)
	}

	c.Set(7, sizerInt(7))

	if evicts != 5 {
		t.Fatal(evicts)
	}
	if l := c.Len(false); l != 7 {
		t.Fatal(l)
	}
}

type sizerInt int

func (s sizerInt) Size() int {
	return int(s)
}
