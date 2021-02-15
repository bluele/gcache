package gcache

import (
	"fmt"
	"testing"
	"time"
)

func TestLFUGet(t *testing.T) {
	size := 1000
	numbers := 1000

	gc := buildTestLoadingCache(t, TYPE_LFU, size, loader)
	testSetCache(t, gc, numbers)
	testGetCache(t, gc, numbers)
}

func TestLoadingLFUGet(t *testing.T) {
	size := 1000
	numbers := 1000

	gc := buildTestLoadingCache(t, TYPE_LFU, size, loader)
	testGetCache(t, gc, numbers)
}

func TestLFULength(t *testing.T) {
	gc := buildTestLoadingCache(t, TYPE_LFU, 1000, loader)
	gc.Get("test1")
	gc.Get("test2")
	length := gc.Len(true)
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", length, expectedLength)
	}
}

func TestLFUEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc := buildTestLoadingCache(t, TYPE_LFU, cacheSize, loader)

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

func TestLFUHas(t *testing.T) {
	gc := buildTestLoadingCacheWithExpiration(t, TYPE_LFU, 2, 10*time.Millisecond)

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

func TestLFUFreqListOrder(t *testing.T) {
	gc := buildTestCache(t, TYPE_LFU, 5)
	for i := 4; i >= 0; i-- {
		gc.Set(i, i)
		for j := 0; j <= i; j++ {
			gc.Get(i)
		}
	}
	if l := gc.(*LFUCache).freqList.Len(); l != 6 {
		t.Fatalf("%v != 6", l)
	}
	var i uint
	for e := gc.(*LFUCache).freqList.Front(); e != nil; e = e.Next() {
		if e.Value.(*freqEntry).freq != i {
			t.Fatalf("%v != %v", e.Value.(*freqEntry).freq, i)
		}
		i++
	}
	gc.Remove(1)

	if l := gc.(*LFUCache).freqList.Len(); l != 5 {
		t.Fatalf("%v != 5", l)
	}
	gc.Set(1, 1)
	if l := gc.(*LFUCache).freqList.Len(); l != 5 {
		t.Fatalf("%v != 5", l)
	}
	gc.Get(1)
	if l := gc.(*LFUCache).freqList.Len(); l != 5 {
		t.Fatalf("%v != 5", l)
	}
	gc.Get(1)
	if l := gc.(*LFUCache).freqList.Len(); l != 6 {
		t.Fatalf("%v != 6", l)
	}
}

func TestLFUFreqListLength(t *testing.T) {
	k0, v0 := "k0", "v0"
	k1, v1 := "k1", "v1"

	{
		gc := buildTestCache(t, TYPE_LFU, 5)
		if l := gc.(*LFUCache).freqList.Len(); l != 1 {
			t.Fatalf("%v != 1", l)
		}
	}
	{
		gc := buildTestCache(t, TYPE_LFU, 5)
		gc.Set(k0, v0)
		for i := 0; i < 5; i++ {
			gc.Get(k0)
		}
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
	}

	{
		gc := buildTestCache(t, TYPE_LFU, 5)
		gc.Set(k0, v0)
		gc.Set(k1, v1)
		for i := 0; i < 5; i++ {
			gc.Get(k0)
			gc.Get(k1)
		}
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
	}

	{
		gc := buildTestCache(t, TYPE_LFU, 5)
		gc.Set(k0, v0)
		gc.Set(k1, v1)
		for i := 0; i < 5; i++ {
			gc.Get(k0)
		}
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
		for i := 0; i < 5; i++ {
			gc.Get(k1)
		}
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
	}

	{
		gc := buildTestCache(t, TYPE_LFU, 5)
		gc.Set(k0, v0)
		gc.Get(k0)
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
		gc.Remove(k0)
		if l := gc.(*LFUCache).freqList.Len(); l != 1 {
			t.Fatalf("%v != 1", l)
		}
		gc.Set(k0, v0)
		if l := gc.(*LFUCache).freqList.Len(); l != 1 {
			t.Fatalf("%v != 1", l)
		}
		gc.Get(k0)
		if l := gc.(*LFUCache).freqList.Len(); l != 2 {
			t.Fatalf("%v != 2", l)
		}
	}
}
