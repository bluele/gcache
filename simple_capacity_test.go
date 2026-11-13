package gcache

import (
	"testing"
	"time"
)

func TestSimpleExpiringEntriesRespectCapacity(t *testing.T) {
	for _, mode := range []string{"default", "per-item", "loader", "loader-per-item"} {
		t.Run(mode, func(t *testing.T) {
			evicted := 0
			builder := New(2).Simple().EvictedFunc(func(key, value interface{}) { evicted++ })
			if mode == "default" || mode == "loader" {
				builder.Expiration(time.Hour)
			}
			if mode == "loader" {
				builder.LoaderFunc(func(key interface{}) (interface{}, error) { return key, nil })
			}
			if mode == "loader-per-item" {
				builder.LoaderExpireFunc(func(key interface{}) (interface{}, *time.Duration, error) { ttl := time.Hour; return key, &ttl, nil })
			}
			cache := builder.Build()
			insert := func(key int) {
				var err error
				switch mode {
				case "per-item":
					err = cache.SetWithExpire(key, key, time.Hour)
				case "loader", "loader-per-item":
					_, err = cache.Get(key)
				default:
					err = cache.Set(key, key)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			insert(0)
			insert(1)
			insert(0)
			if evicted != 0 {
				t.Fatal("updating an existing key evicted an item")
			}
			for key := 2; key < 100; key++ {
				insert(key)
				if got := cache.Len(false); got != 2 {
					t.Fatalf("after key %d, stored %d entries; want 2", key, got)
				}
				if value, err := cache.GetIFPresent(key); err != nil || value != key {
					t.Fatalf("new entry is unavailable: %v, %v", value, err)
				}
				if evicted != key-1 {
					t.Fatalf("after key %d, got %d evictions, want %d", key, evicted, key-1)
				}
			}
		})
	}
}

func TestSimpleCapacityPrefersExpiredEntries(t *testing.T) {
	clock := NewFakeClock()
	var evicted []interface{}
	cache := New(2).Clock(clock).EvictedFunc(func(key, value interface{}) { evicted = append(evicted, key) }).Build()
	if err := cache.SetWithExpire("expired", "old", time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := cache.SetWithExpire("live", "value", time.Hour); err != nil {
		t.Fatal(err)
	}
	clock.Advance(2 * time.Minute)
	if err := cache.SetWithExpire("new", "new-value", time.Hour); err != nil {
		t.Fatal(err)
	}
	if len(evicted) != 1 || evicted[0] != "expired" {
		t.Fatalf("evicted %v, want expired", evicted)
	}
	for _, key := range []string{"live", "new"} {
		if _, err := cache.Get(key); err != nil {
			t.Fatalf("lost live key %s: %v", key, err)
		}
	}
}

func TestSimpleExpiringEntriesRemainUnbounded(t *testing.T) {
	for _, size := range []int{0, -1} {
		cache := New(size).Expiration(time.Hour).Build()
		for key := 0; key < 100; key++ {
			if err := cache.Set(key, key); err != nil {
				t.Fatal(err)
			}
		}
		if got := cache.Len(false); got != 100 {
			t.Fatalf("size %d retained %d entries, want 100", size, got)
		}
	}
}
