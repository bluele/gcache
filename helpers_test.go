package gcache

import (
	"fmt"
	"testing"
	"time"
)

func loader(key interface{}) (interface{}, error) {
	return fmt.Sprintf("valueFor%s", key), nil
}

func testSetCache(t *testing.T, gc Cache, numbers int) {
	for i := 0; i < numbers; i++ {
		key := fmt.Sprintf("Key-%d", i)
		value, err := loader(key)
		if err != nil {
			t.Error(err)
			return
		}
		gc.Set(key, value)
	}
}

func testGetCache(t *testing.T, gc Cache, numbers int) {
	for i := 0; i < numbers; i++ {
		key := fmt.Sprintf("Key-%d", i)
		v, err := gc.Get(key)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expectedV, _ := loader(key)
		if v != expectedV {
			t.Errorf("Expected value is %v, not %v", expectedV, v)
		}
	}
}

func testGetIFPresent(t *testing.T, evT string) {
	cache :=
		New(8).
			EvictType(evT).
			LoaderFunc(
				func(key interface{}) (interface{}, error) {
					return "value", nil
				}).
			Build()

	v, err := cache.GetIFPresent("key")
	if err != KeyNotFoundError {
		t.Errorf("err should not be %v", err)
	}

	time.Sleep(2 * time.Millisecond)

	v, err = cache.GetIFPresent("key")
	if err != nil {
		t.Errorf("err should not be %v", err)
	}
	if v != "value" {
		t.Errorf("v should not be %v", v)
	}
}

func testGetALL(t *testing.T, evT string) {
	size := 8
	cache :=
		New(size).
			Expiration(time.Millisecond).
			EvictType(evT).
			Build()
	for i := 0; i < size; i++ {
		cache.Set(i, i*i)
	}
	m := cache.GetALL()
	for i := 0; i < size; i++ {
		v, ok := m[i]
		if !ok {
			t.Errorf("m should contain %v", i)
			continue
		}
		if v.(int) != i*i {
			t.Errorf("%v != %v", v, i*i)
			continue
		}
	}
	time.Sleep(time.Millisecond)

	cache.Set(size, size*size)
	m = cache.GetALL()
	if len(m) != 1 {
		t.Errorf("%v != %v", len(m), 1)
	}
	if _, ok := m[size]; !ok {
		t.Errorf("%v should contains key '%v'", m, size)
	}
}
