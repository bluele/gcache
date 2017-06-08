package gcache

import (
	"bytes"
	"encoding/gob"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoaderFunc(t *testing.T) {
	size := 2
	var testCaches = []*CacheBuilder{
		New(size).Simple(),
		New(size).LRU(),
		New(size).LFU(),
		New(size).ARC(),
	}
	for _, builder := range testCaches {
		var testCounter int64
		counter := 1000
		cache := builder.
			LoaderFunc(func(key interface{}) (interface{}, error) {
				time.Sleep(10 * time.Millisecond)
				return atomic.AddInt64(&testCounter, 1), nil
			}).
			EvictedFunc(func(key, value interface{}) {
				panic(key)
			}).Build()

		var wg sync.WaitGroup
		for i := 0; i < counter; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := cache.Get(0)
				if err != nil {
					t.Error(err)
				}
			}()
		}
		wg.Wait()

		if testCounter != 1 {
			t.Errorf("testCounter != %v", testCounter)
		}
	}
}

func TestLoaderExpireFuncWithoutExpire(t *testing.T) {
	size := 2
	var testCaches = []*CacheBuilder{
		New(size).Simple(),
		New(size).LRU(),
		New(size).LFU(),
		New(size).ARC(),
	}
	for _, builder := range testCaches {
		var testCounter int64
		counter := 1000
		cache := builder.
			LoaderExpireFunc(func(key interface{}) (interface{}, *time.Duration, error) {
				return atomic.AddInt64(&testCounter, 1), nil, nil
			}).
			EvictedFunc(func(key, value interface{}) {
				panic(key)
			}).Build()

		var wg sync.WaitGroup
		for i := 0; i < counter; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := cache.Get(0)
				if err != nil {
					t.Error(err)
				}
			}()
		}

		wg.Wait()

		if testCounter != 1 {
			t.Errorf("testCounter != %v", testCounter)
		}
	}
}

func TestLoaderExpireFuncWithExpire(t *testing.T) {
	size := 2
	var testCaches = []*CacheBuilder{
		New(size).Simple(),
		New(size).LRU(),
		New(size).LFU(),
		New(size).ARC(),
	}
	for _, builder := range testCaches {
		var testCounter int64
		counter := 1000
		expire := 200 * time.Millisecond
		cache := builder.
			LoaderExpireFunc(func(key interface{}) (interface{}, *time.Duration, error) {
				return atomic.AddInt64(&testCounter, 1), &expire, nil
			}).
			Build()

		var wg sync.WaitGroup
		for i := 0; i < counter; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := cache.Get(0)
				if err != nil {
					t.Error(err)
				}
			}()
		}
		time.Sleep(expire) // Waiting for key expiration
		for i := 0; i < counter; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := cache.Get(0)
				if err != nil {
					t.Error(err)
				}
			}()
		}

		wg.Wait()

		if testCounter != 2 {
			t.Errorf("testCounter != %v", testCounter)
		}
	}
}

func TestDeserializeFunc(t *testing.T) {
	var cases = []struct {
		tp string
	}{
		{TYPE_SIMPLE},
		{TYPE_LRU},
		{TYPE_LFU},
		{TYPE_ARC},
	}

	for _, cs := range cases {
		key1, value1 := "key1", "value1"
		key2, value2 := "key2", "value2"
		cc := New(32).
			EvictType(cs.tp).
			LoaderFunc(func(k interface{}) (interface{}, error) {
				return value1, nil
			}).
			DeserializeFunc(func(k, v interface{}) (interface{}, error) {
				dec := gob.NewDecoder(bytes.NewBuffer(v.([]byte)))
				var str string
				err := dec.Decode(&str)
				if err != nil {
					return nil, err
				}
				return str, nil
			}).
			SerializeFunc(func(k, v interface{}) (interface{}, error) {
				buf := new(bytes.Buffer)
				enc := gob.NewEncoder(buf)
				err := enc.Encode(v)
				return buf.Bytes(), err
			}).
			Build()
		v, err := cc.Get(key1)
		if err != nil {
			t.Fatal(err)
		}
		if v != value1 {
			t.Errorf("%v != %v", v, value1)
		}
		v, err = cc.Get(key1)
		if err != nil {
			t.Fatal(err)
		}
		if v != value1 {
			t.Errorf("%v != %v", v, value1)
		}
		if err := cc.Set(key2, value2); err != nil {
			t.Error(err)
		}
		v, err = cc.Get(key2)
		if err != nil {
			t.Error(err)
		}
		if v != value2 {
			t.Errorf("%v != %v", v, value2)
		}
	}
}
