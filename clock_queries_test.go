package gcache

import (
	"testing"
	"time"
)

func TestQueriesUseConfiguredClock(t *testing.T) {
	for _, typ := range []string{TYPE_SIMPLE, TYPE_LRU, TYPE_LFU, TYPE_ARC} {
		for _, epoch := range []int{1984, 2100} {
			for _, perItem := range []bool{false, true} {
				t.Run(typ+"/"+time.Date(epoch, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")+"/"+map[bool]string{false: "default", true: "per-item"}[perItem], func(t *testing.T) {
					clock := &fakeclock{now: time.Date(epoch, 1, 1, 0, 0, 0, 0, time.UTC)}
					cache := New(4).EvictType(typ).Clock(clock).Expiration(time.Hour).Build()
					var err error
					if perItem {
						err = cache.SetWithExpire("key", "value", time.Hour)
					} else {
						err = cache.Set("key", "value")
					}
					if err != nil {
						t.Fatal(err)
					}
					for _, expired := range []bool{false, true} {
						if expired {
							clock.Advance(time.Hour + time.Nanosecond)
						}
						want := !expired
						if got := cache.Has("key"); got != want {
							t.Errorf("expired=%v Has=%v, want %v", expired, got, want)
						}
						if got := cache.Len(true); (got == 1) != want {
							t.Errorf("expired=%v Len(true)=%d", expired, got)
						}
						if got := cache.Keys(true); (len(got) == 1) != want {
							t.Errorf("expired=%v Keys(true)=%v", expired, got)
						}
						if got := cache.GetALL(true); (len(got) == 1) != want {
							t.Errorf("expired=%v GetALL(true)=%v", expired, got)
						}
						if cache.Len(false) != 1 || len(cache.Keys(false)) != 1 || len(cache.GetALL(false)) != 1 {
							t.Error("unfiltered queries lost stored item")
						}
						// Get already uses the configured clock, so query APIs must agree with it.
						value, err := cache.Get("key")
						if want && (err != nil || value != "value") {
							t.Errorf("Get=%v,%v, want value,nil", value, err)
						}
						if expired && err != KeyNotFoundError {
							t.Errorf("expired Get returned %v", err)
						}
					}
				})
			}
		}
	}
}
