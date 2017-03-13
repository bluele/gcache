package gcache_test

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	"sync"
	"sync/atomic"

	"github.com/bluele/gcache"
)

func TestLoaderFunc(t *testing.T) {
	size := 2
	var testCaches = []*gcache.CacheBuilder{
		gcache.New(size).Simple(),
		gcache.New(size).LRU(),
		gcache.New(size).LFU(),
		gcache.New(size).ARC(),
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

func TestGetterFunc(t *testing.T) {
	var cases = []struct {
		tp string
	}{
		{gcache.TYPE_SIMPLE},
		{gcache.TYPE_LRU},
		{gcache.TYPE_LFU},
		{gcache.TYPE_ARC},
	}

	for _, cs := range cases {
		key1, value1 := "key1", "value1"
		key2, value2 := "key2", "value2"
		cc := gcache.New(32).
			EvictType(cs.tp).
			LoaderFunc(func(k interface{}) (interface{}, error) {
				return value1, nil
			}).
			GetterFunc(func(k, v interface{}) (interface{}, error) {
				dec := gob.NewDecoder(bytes.NewBuffer(v.([]byte)))
				var str string
				err := dec.Decode(&str)
				if err != nil {
					return nil, err
				}
				return str, nil
			}).
			SetterFunc(func(k, v interface{}) (interface{}, error) {
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
