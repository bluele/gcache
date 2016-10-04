package gcache

import (
	"testing"
)

func TestStats(t *testing.T) {
	var cases = []struct {
		hit  int
		miss int
		rate float64
	}{
		{3, 1, 0.75},
		{0, 1, 0.0},
		{3, 0, 1.0},
		{0, 0, 0.0},
	}

	for _, cs := range cases {
		st := &stats{}
		for i := 0; i < cs.hit; i++ {
			st.IncrHitCount()
		}
		for i := 0; i < cs.miss; i++ {
			st.IncrMissCount()
		}
		if rate := st.HitRate(); rate != cs.rate {
			t.Errorf("%v != %v", rate, cs.rate)
		}
	}
}

func getter(key interface{}) (interface{}, error) {
	return key, nil
}

func TestCacheStats(t *testing.T) {
	var cases = []struct {
		builder func() Cache
		rate    float64
	}{
		{
			builder: func() Cache {
				cc := New(32).Simple().Build()
				cc.Set(0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() Cache {
				cc := New(32).LRU().Build()
				cc.Set(0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() Cache {
				cc := New(32).LFU().Build()
				cc.Set(0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() Cache {
				cc := New(32).ARC().Build()
				cc.Set(0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() Cache {
				cc := New(32).
					Simple().
					LoaderFunc(getter).
					Build()
				cc.Set(0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() Cache {
				cc := New(32).
					LRU().
					LoaderFunc(getter).
					Build()
				cc.Set(0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() Cache {
				cc := New(32).
					LFU().
					LoaderFunc(getter).
					Build()
				cc.Set(0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
		{
			builder: func() Cache {
				cc := New(32).
					ARC().
					LoaderFunc(getter).
					Build()
				cc.Set(0, 0)
				cc.Get(0)
				cc.Get(1)
				return cc
			},
			rate: 0.5,
		},
	}

	for i, cs := range cases {
		cc := cs.builder()
		if rate := cc.HitRate(); rate != cs.rate {
			t.Errorf("case-%v: %v != %v", i, rate, cs.rate)
		}
	}
}
