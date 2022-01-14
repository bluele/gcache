package gcache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	TYPE_SIMPLE = "simple"
	TYPE_LRU    = "lru"
	TYPE_LFU    = "lfu"
	TYPE_ARC    = "arc"
)

var KeyNotFoundError = errors.New("Key not found.")

type Cache interface {
	// Set inserts or updates the specified key-value pair.
	Set(key, value interface{}) error
	// SetWithExpire inserts or updates the specified key-value pair with an expiration time.
	SetWithExpire(key, value interface{}, expiration time.Duration) error
	// Get returns the value for the specified key if it is present in the cache.
	// If the key is not present in the cache and the cache has LoaderFunc,
	// invoke the `LoaderFunc` function and inserts the key-value pair in the cache.
	// If the key is not present in the cache and the cache does not have a LoaderFunc,
	// return KeyNotFoundError.
	Get(key interface{}) (interface{}, error)
	// GetIFPresent returns the value for the specified key if it is present in the cache.
	// Return KeyNotFoundError if the key is not present.
	GetIFPresent(key interface{}) (interface{}, error)
	// GetAll returns a map containing all key-value pairs in the cache.
	GetALL(checkExpired bool) map[interface{}]interface{}
	get(key interface{}, onLoad bool) (interface{}, error)
	// Remove removes the specified key from the cache if the key is present.
	// Returns true if the key was present and the key has been deleted.
	Remove(key interface{}) bool
	// Purge removes all key-value pairs from the cache.
	Purge()
	// Keys returns a slice containing all keys in the cache.
	Keys(checkExpired bool) []interface{}
	// Len returns the number of items in the cache.
	Len(checkExpired bool) int
	// Has returns true if the key exists in the cache.
	Has(key interface{}) bool
	GetContext(ctx context.Context, key interface{}) (interface{}, error)
	GetIFPresentContext(ctx context.Context, key interface{}) (interface{}, error)

	statsAccessor
}

type baseCache struct {
	clock                   Clock
	size                    int
	loaderExpireContextFunc LoaderExpireContextFunc
	evictedFunc             EvictedFunc
	purgeVisitorFunc        PurgeVisitorFunc
	addedFunc               AddedFunc
	deserializeFunc         DeserializeFunc
	serializeFunc           SerializeFunc
	expiration              *time.Duration
	mu                      sync.RWMutex
	loadGroup               Group
	*stats
}

type (
	LoaderFunc              func(interface{}) (interface{}, error)
	LoaderContextFunc       func(context.Context, interface{}) (interface{}, error)
	LoaderExpireFunc        func(interface{}) (interface{}, *time.Duration, error)
	LoaderExpireContextFunc func(context.Context, interface{}) (interface{}, *time.Duration, error)
	EvictedFunc             func(interface{}, interface{})
	PurgeVisitorFunc        func(interface{}, interface{})
	AddedFunc               func(interface{}, interface{})
	DeserializeFunc         func(interface{}, interface{}) (interface{}, error)
	SerializeFunc           func(interface{}, interface{}) (interface{}, error)
)

type CacheBuilder struct {
	clock                   Clock
	tp                      string
	size                    int
	loaderExpireContextFunc LoaderExpireContextFunc
	evictedFunc             EvictedFunc
	purgeVisitorFunc        PurgeVisitorFunc
	addedFunc               AddedFunc
	expiration              *time.Duration
	deserializeFunc         DeserializeFunc
	serializeFunc           SerializeFunc
}

func New(size int) *CacheBuilder {
	return &CacheBuilder{
		clock: NewRealClock(),
		tp:    TYPE_SIMPLE,
		size:  size,
	}
}

func (cb *CacheBuilder) Clock(clock Clock) *CacheBuilder {
	cb.clock = clock
	return cb
}

// Set a loader function.
// loaderFunc: create a new value with this function if cached value is expired.
func (cb *CacheBuilder) LoaderFunc(loaderFunc LoaderFunc) *CacheBuilder {
	cb.loaderExpireContextFunc = func(_ context.Context, k interface{}) (interface{}, *time.Duration, error) {
		v, err := loaderFunc(k)
		return v, nil, err
	}
	return cb
}

func (cb *CacheBuilder) LoaderContextFunc(loaderContextFunc LoaderContextFunc) *CacheBuilder {
	cb.loaderExpireContextFunc = func(ctx context.Context, k interface{}) (interface{}, *time.Duration, error) {
		v, err := loaderContextFunc(ctx, k)
		return v, nil, err
	}
	return cb
}

// Set a loader function with expiration.
// loaderExpireContextFunc: create a new value with this function if cached value is expired.
// If nil returned instead of time.Duration from loaderExpireContextFunc than value will never expire.
func (cb *CacheBuilder) LoaderExpireFunc(loaderExpireFunc LoaderExpireFunc) *CacheBuilder {
	cb.loaderExpireContextFunc = func(_ context.Context, i2 interface{}) (i interface{}, duration *time.Duration, err error) {
		return loaderExpireFunc(i2)
	}
	return cb
}

func (cb *CacheBuilder) LoaderExpireContextFunc(loaderExpireContextFunc LoaderExpireContextFunc) *CacheBuilder {
	cb.loaderExpireContextFunc = loaderExpireContextFunc
	return cb
}

func (cb *CacheBuilder) EvictType(tp string) *CacheBuilder {
	cb.tp = tp
	return cb
}

func (cb *CacheBuilder) Simple() *CacheBuilder {
	return cb.EvictType(TYPE_SIMPLE)
}

func (cb *CacheBuilder) LRU() *CacheBuilder {
	return cb.EvictType(TYPE_LRU)
}

func (cb *CacheBuilder) LFU() *CacheBuilder {
	return cb.EvictType(TYPE_LFU)
}

func (cb *CacheBuilder) ARC() *CacheBuilder {
	return cb.EvictType(TYPE_ARC)
}

func (cb *CacheBuilder) EvictedFunc(evictedFunc EvictedFunc) *CacheBuilder {
	cb.evictedFunc = evictedFunc
	return cb
}

func (cb *CacheBuilder) PurgeVisitorFunc(purgeVisitorFunc PurgeVisitorFunc) *CacheBuilder {
	cb.purgeVisitorFunc = purgeVisitorFunc
	return cb
}

func (cb *CacheBuilder) AddedFunc(addedFunc AddedFunc) *CacheBuilder {
	cb.addedFunc = addedFunc
	return cb
}

func (cb *CacheBuilder) DeserializeFunc(deserializeFunc DeserializeFunc) *CacheBuilder {
	cb.deserializeFunc = deserializeFunc
	return cb
}

func (cb *CacheBuilder) SerializeFunc(serializeFunc SerializeFunc) *CacheBuilder {
	cb.serializeFunc = serializeFunc
	return cb
}

func (cb *CacheBuilder) Expiration(expiration time.Duration) *CacheBuilder {
	cb.expiration = &expiration
	return cb
}

func (cb *CacheBuilder) Build() Cache {
	if cb.size <= 0 && cb.tp != TYPE_SIMPLE {
		panic("gcache: Cache size <= 0")
	}

	return cb.build()
}

func (cb *CacheBuilder) build() Cache {
	switch cb.tp {
	case TYPE_SIMPLE:
		return newSimpleCache(cb)
	case TYPE_LRU:
		return newLRUCache(cb)
	case TYPE_LFU:
		return newLFUCache(cb)
	case TYPE_ARC:
		return newARC(cb)
	default:
		panic("gcache: Unknown type " + cb.tp)
	}
}

func buildCache(c *baseCache, cb *CacheBuilder) {
	c.clock = cb.clock
	c.size = cb.size
	c.loaderExpireContextFunc = cb.loaderExpireContextFunc
	c.expiration = cb.expiration
	c.addedFunc = cb.addedFunc
	c.deserializeFunc = cb.deserializeFunc
	c.serializeFunc = cb.serializeFunc
	c.evictedFunc = cb.evictedFunc
	c.purgeVisitorFunc = cb.purgeVisitorFunc
	c.stats = &stats{}
}

// load a new value using by specified key.
func (c *baseCache) load(ctx context.Context, key interface{}, cb func(interface{}, *time.Duration, error) (interface{}, error), isWait bool) (interface{}, bool, error) {
	v, called, err := c.loadGroup.Do(key, func() (v interface{}, e error) {
		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("Loader panics: %v", r)
			}
		}()
		return cb(c.loaderExpireContextFunc(ctx, key))
	}, isWait)
	if err != nil {
		return nil, called, err
	}
	return v, called, nil
}
