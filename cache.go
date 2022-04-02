package gcache

import (
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

	statsAccessor
}

type baseCache struct {
	clock            Clock
	size             int
	loaderExpireFunc LoaderExpireFunc
	evictedFunc      EvictedFunc
	purgeVisitorFunc PurgeVisitorFunc
	addedFunc        AddedFunc
	deserializeFunc  DeserializeFunc
	serializeFunc    SerializeFunc
	expiration       *time.Duration
	mu               Locker
	loadGroup        Group
	*stats
}

type (
	LoaderFunc       func(interface{}) (interface{}, error)
	LoaderExpireFunc func(interface{}) (interface{}, *time.Duration, error)
	EvictedFunc      func(interface{}, interface{})
	PurgeVisitorFunc func(interface{}, interface{})
	AddedFunc        func(interface{}, interface{})
	DeserializeFunc  func(interface{}, interface{}) (interface{}, error)
	SerializeFunc    func(interface{}, interface{}) (interface{}, error)
)

type Locker interface {
	RLock()
	RUnlock()
	sync.Locker
}

type noopLocker struct{}

func (noopLocker) RLock()   {}
func (noopLocker) RUnlock() {}
func (noopLocker) Lock()    {}
func (noopLocker) Unlock()  {}

type CacheBuilder struct {
	clock            Clock
	tp               string
	size             int
	loaderExpireFunc LoaderExpireFunc
	evictedFunc      EvictedFunc
	purgeVisitorFunc PurgeVisitorFunc
	addedFunc        AddedFunc
	expiration       *time.Duration
	deserializeFunc  DeserializeFunc
	serializeFunc    SerializeFunc
	locker           Locker
}

func New(size int) *CacheBuilder {
	return &CacheBuilder{
		clock:  NewRealClock(),
		tp:     TYPE_SIMPLE,
		size:   size,
		locker: &sync.RWMutex{},
	}
}

func (cb *CacheBuilder) NoLock() *CacheBuilder {
	cb.locker = noopLocker{}
	return cb
}

func (cb *CacheBuilder) Clock(clock Clock) *CacheBuilder {
	cb.clock = clock
	return cb
}

// Set a loader function.
// loaderFunc: create a new value with this function if cached value is expired.
func (cb *CacheBuilder) LoaderFunc(loaderFunc LoaderFunc) *CacheBuilder {
	cb.loaderExpireFunc = func(k interface{}) (interface{}, *time.Duration, error) {
		v, err := loaderFunc(k)
		return v, nil, err
	}
	return cb
}

// Set a loader function with expiration.
// loaderExpireFunc: create a new value with this function if cached value is expired.
// If nil returned instead of time.Duration from loaderExpireFunc than value will never expire.
func (cb *CacheBuilder) LoaderExpireFunc(loaderExpireFunc LoaderExpireFunc) *CacheBuilder {
	cb.loaderExpireFunc = loaderExpireFunc
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
	c.loaderExpireFunc = cb.loaderExpireFunc
	c.expiration = cb.expiration
	c.addedFunc = cb.addedFunc
	c.deserializeFunc = cb.deserializeFunc
	c.serializeFunc = cb.serializeFunc
	c.evictedFunc = cb.evictedFunc
	c.purgeVisitorFunc = cb.purgeVisitorFunc
	c.stats = &stats{}
	c.mu = cb.locker
}

// load a new value using by specified key.
func (c *baseCache) load(key interface{}, cb func(interface{}, *time.Duration, error) (interface{}, error), isWait bool) (interface{}, bool, error) {
	v, called, err := c.loadGroup.Do(key, func() (v interface{}, e error) {
		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("Loader panics: %v", r)
			}
		}()
		return cb(c.loaderExpireFunc(key))
	}, isWait)
	if err != nil {
		return nil, called, err
	}
	return v, called, nil
}
