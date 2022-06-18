package gcache

import (
	"fmt"
	"testing"
	"time"
)

func TestARCShardGet(t *testing.T) {
	size := 10
	gc := buildTest(t, TYPE_ARC, size).ShardCount(3).Build()
	//gc.Set("k1", "v1")

	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

//func TestLoadingARCShardGet(t *testing.T) {
//	size := 1000
//	numbers := 1000
//	testGetCache(t, buildTestLoadingCache(t, TYPE_ARC, size, loader), numbers)
//}

func TestARCShardLength(t *testing.T) {
	shardCount := 3
	gc := buildTestLoadingWithExpiration(t, TYPE_ARC, 2, time.Millisecond).ShardCount(shardCount).Build()
	gc.Get("test1")
	gc.Get("test2")
	gc.Get("test3")
	length := gc.Len(true)
	expectedLength := 2
	if length < shardCount || length > shardCount*expectedLength {
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

func TestARCShardEvictItem(t *testing.T) {
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

func TestARCShardPurgeCache(t *testing.T) {
	cacheSize := 10
	purgeCount := 0
	gc := New(cacheSize).
		ARC().
		LoaderFunc(loader).
		PurgeVisitorFunc(func(k, v interface{}) {
			purgeCount++
		}).
		ShardCount(3).
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

//
//func TestARCShardGetIFPresent(t *testing.T) {
//	testGetIFPresent(t, TYPE_ARC)
//}
//
func TestARCShardHas(t *testing.T) {
	gc := buildTestLoadingWithExpiration(t, TYPE_ARC, 2, 10*time.Millisecond).ShardCount(3).Build()

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
