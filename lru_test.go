package gcache

import (
	"fmt"
	"testing"
	"time"
)

func TestLRUGet(t *testing.T) {
	size := 1000
	gc := buildTestCache(t, TYPE_LRU, size)
	testSetCache(t, gc, size)
	testGetCache(t, gc, size)
}

func TestLoadingLRUGet(t *testing.T) {
	size := 1000
	gc := buildTestLoadingCache(t, TYPE_LRU, size, loader)
	testGetCache(t, gc, size)
}

func TestLRULength(t *testing.T) {
	gc := buildTestLoadingCache(t, TYPE_LRU, 1000, loader)
	_, _ = gc.Get("test1")
	_, _ = gc.Get("test2")
	length := gc.Len(true)
	expectedLength := 2
	if length != expectedLength {
		t.Errorf("Expected length is %v, not %v", length, expectedLength)
	}
}

func TestLRUEvictItem(t *testing.T) {
	cacheSize := 10
	numbers := 11
	gc := buildTestLoadingCache(t, TYPE_LRU, cacheSize, loader)

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

func TestLRUHas(t *testing.T) {
	gc := buildTestLoadingCacheWithExpiration(t, TYPE_LRU, 2, 10*time.Millisecond)

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			_, _ = gc.Get("test1")
			_, _ = gc.Get("test2")

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

func TestBasicLRUIncrementer(t *testing.T) {
	gc := buildTestLoadingCacheWithExpiration(t, TYPE_LRU, 100, 10*time.Second)
	defer gc.Purge()

	// integer
	err := gc.Set("some-key", 1)
	if err != nil {
		t.Error(err)
		t.Fatal()
	}
	v, err := gc.Increment("some-key", 1)
	if err != nil {
		t.Error(err)
		t.Fatal()
	}
	if v == nil {
		t.Error(fmt.Errorf("v is nil"))
		t.Fatal()
	}
	vNew, ok := v.(int)
	if !ok {
		t.Error(fmt.Errorf("vNew is not int"))
		t.Fatal()
	}
	if vNew != 2 {
		t.Error("increment int failed")
		t.Fatal()
	}
	vFromC, err := gc.Get("some-key")
	if err != nil {
		t.Error(err)
		t.Fatal()
	}
	if vFromC != vNew {
		t.Error(fmt.Errorf("increment in cache by int64 failed, v:%v, vNew:%v", vNew, vFromC))
	}
}

func TestLRUIncrementer(t *testing.T) {
	gc := buildTestLoadingCacheWithExpiration(t, TYPE_LRU, 100, 10*time.Second)
	defer gc.Purge()
	featureValues := []interface{}{
		[]int{1, 2},
		[]int8{1, 0, 5},
		[]int16{1, 0, 5},
		[]int32{1, 0, 5},
		[]int64{1, 0, 5},
		[]uint{1, 0, 5},
		[]uintptr{1, 0, 5},
		[]uint8{1, 0, 5},
		[]uint16{1, 0, 5},
		[]uint32{1, 0, 5},
		[]uint64{1, 0, 5},
		[]float32{1.5, 0.2, 5.4},
		[]float64{1.5, 0.7, 5.9},
	}

	for _, values := range featureValues {
		forEachValue(values, func(i int, v interface{}) {
			incrementBy := int64(1)
			var incrementResult interface{}

			switch v.(type) {
			case int:
				v = v.(int)
				incrementResult = int(incrementBy) + v.(int)
			case int8:
				v = v.(int8)
				incrementResult = int8(incrementBy) + v.(int8)
			case int16:
				v = v.(int16)
				incrementResult = int16(incrementBy) + v.(int16)
			case int32:
				v = v.(int32)
				incrementResult = int32(incrementBy) + v.(int32)
			case int64:
				v = v.(int64)
				incrementResult = int64(incrementBy) + v.(int64)
			case uint:
				v = v.(uint)
				incrementResult = uint(incrementBy) + v.(uint)
			case uintptr:
				v = v.(uintptr)
				incrementResult = uintptr(incrementBy) + v.(uintptr)
			case uint8:
				v = v.(uint8)
				incrementResult = uint8(incrementBy) + v.(uint8)
			case uint16:
				v = v.(uint16)
				incrementResult = uint16(incrementBy) + v.(uint16)
			case uint32:
				v = v.(uint32)
				incrementResult = uint32(incrementBy) + v.(uint32)
			case uint64:
				v = v.(uint64)
				incrementResult = uint64(incrementBy) + v.(uint64)
			case float32:
				v = v.(float32)
				incrementResult = float32(incrementBy) + v.(float32)
			case float64:
				v = v.(float64)
				incrementResult = float64(incrementBy) + v.(float64)
			default:
				t.Error(fmt.Errorf("the value %v is not an integer", v))
				t.Fatal()
			}

			err := gc.Set("some-key", v)
			if err != nil {
				t.Error(err)
				t.Fatal()
			}

			vNew, err := gc.Increment("some-key", incrementBy)
			if err != nil {
				t.Error(err)
				t.Fatal()
			}

			if vNew == nil {
				t.Error(fmt.Errorf("v is nil"))
				t.Fatal()
			}

			switch vNew.(type) {
			case int:
				vNew = vNew.(int)
			case int8:
				vNew = vNew.(int8)
			case int16:
				vNew = vNew.(int16)
			case int32:
				vNew = vNew.(int32)
			case int64:
				vNew = vNew.(int64)
			case uint:
				vNew = vNew.(uint)
			case uintptr:
				vNew = vNew.(uintptr)
			case uint8:
				vNew = vNew.(uint8)
			case uint16:
				vNew = vNew.(uint16)
			case uint32:
				vNew = vNew.(uint32)
			case uint64:
				vNew = vNew.(uint64)
			case float32:
				vNew = vNew.(float32)
			case float64:
				vNew = vNew.(float64)
			default:
				t.Error(fmt.Errorf("the value %v is not an integer", vNew))
				t.Fatal()
			}

			if vNew != incrementResult {
				t.Error(fmt.Errorf("increment result by int64 failed, v:%v, vNew:%v", v, vNew))
				//t.Fatal()
			} else {
				t.Logf("value:%v, incremented:%v, by number: %d \n", v, vNew, incrementBy)
			}

			vFromC, err := gc.Get("some-key")
			if err != nil {
				t.Error(err)
				t.Fatal()
			}
			// ugly hack to compare different types
			if vFromC != vNew {
				t.Error(fmt.Errorf("increment in cache by int64 failed, v:%v, vNew:%v", vNew, vFromC))
			}
		})
	}
}
