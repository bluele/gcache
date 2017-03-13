package gcache

import "time"

// SimpleCache has no clear priority for evict cache. It depends on key-value map order.
type SimpleCache struct {
	baseCache
	items map[interface{}]*simpleItem
}

func newSimpleCache(cb *CacheBuilder) *SimpleCache {
	c := &SimpleCache{}
	buildCache(&c.baseCache, cb)

	c.init()
	c.loadGroup.cache = c
	return c
}

func (c *SimpleCache) init() {
	c.items = make(map[interface{}]*simpleItem, c.size)
}

// set a new key-value pair
func (c *SimpleCache) Set(key, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.set(key, value)
	return err
}

func (c *SimpleCache) set(key, value interface{}) (interface{}, error) {
	var err error
	if c.setterFunc != nil {
		value, err = c.setterFunc(key, value)
		if err != nil {
			return nil, err
		}
	}

	// Check for existing item
	item, ok := c.items[key]
	if ok {
		item.value = value
	} else {
		// Verify size not exceeded
		if len(c.items) >= c.size {
			c.evict(1)
		}
		item = &simpleItem{
			value: value,
		}
		c.items[key] = item
	}

	if c.expiration != nil {
		t := time.Now().Add(*c.expiration)
		item.expiration = &t
	}

	if c.addedFunc != nil {
		c.addedFunc(key, value)
	}

	return item, nil
}

// Get a value from cache pool using key if it exists.
// If it dose not exists key and has LoaderFunc,
// generate a value using `LoaderFunc` method returns value.
func (c *SimpleCache) Get(key interface{}) (interface{}, error) {
	v, err := c.getValue(key)
	if err == KeyNotFoundError {
		return c.getWithLoader(key, true)
	}
	return v, err
}

// Get a value from cache pool using key if it exists.
// If it dose not exists key, returns KeyNotFoundError.
// And send a request which refresh value for specified key if cache object has LoaderFunc.
func (c *SimpleCache) GetIFPresent(key interface{}) (interface{}, error) {
	v, err := c.getValue(key)
	if err == KeyNotFoundError {
		return c.getWithLoader(key, false)
	}
	return v, nil
}

func (c *SimpleCache) get(key interface{}, onLoad bool) (interface{}, error) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if ok {
		if !item.IsExpired(nil) {
			if !onLoad {
				c.stats.IncrHitCount()
			}
			return item, nil
		}
		c.mu.Lock()
		c.remove(key)
		c.mu.Unlock()
	}
	if !onLoad {
		c.stats.IncrMissCount()
	}
	return nil, KeyNotFoundError
}

func (c *SimpleCache) getValue(key interface{}) (interface{}, error) {
	it, err := c.get(key, false)
	if err != nil {
		return nil, err
	}
	v := it.(*simpleItem).value
	if c.getterFunc != nil {
		return c.getterFunc(key, v)
	}
	return v, nil
}

func (c *SimpleCache) getWithLoader(key interface{}, isWait bool) (interface{}, error) {
	if c.loaderFunc == nil {
		return nil, KeyNotFoundError
	}
	value, _, err := c.load(key, func(v interface{}, e error) (interface{}, error) {
		if e != nil {
			return nil, e
		}
		c.mu.Lock()
		it, err := c.set(key, v)
		if err != nil {
			c.mu.Unlock()
			return nil, err
		}
		v = it.(*simpleItem).value
		if c.getterFunc == nil {
			c.mu.Unlock()
			return v, nil
		}
		c.mu.Unlock()
		return c.getterFunc(key, v)
	}, isWait)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (c *SimpleCache) evict(count int) {
	now := time.Now()
	current := 0
	for key, item := range c.items {
		if current >= count {
			return
		}
		if item.expiration == nil || now.After(*item.expiration) {
			defer c.remove(key)
			current += 1
		}
	}
}

// Removes the provided key from the cache.
func (c *SimpleCache) Remove(key interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *SimpleCache) remove(key interface{}) bool {
	item, ok := c.items[key]
	if ok {
		delete(c.items, key)
		if c.evictedFunc != nil {
			c.evictedFunc(key, item.value)
		}
		return true
	}
	return false
}

// Returns a slice of the keys in the cache.
func (c *SimpleCache) keys() []interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]interface{}, len(c.items))
	var i = 0
	for k := range c.items {
		keys[i] = k
		i++
	}
	return keys
}

// Returns a slice of the keys in the cache.
func (c *SimpleCache) Keys() []interface{} {
	keys := []interface{}{}
	for _, k := range c.keys() {
		_, err := c.GetIFPresent(k)
		if err == nil {
			keys = append(keys, k)
		}
	}
	return keys
}

// Returns all key-value pairs in the cache.
func (c *SimpleCache) GetALL() map[interface{}]interface{} {
	m := make(map[interface{}]interface{})
	for _, k := range c.keys() {
		v, err := c.GetIFPresent(k)
		if err == nil {
			m[k] = v
		}
	}
	return m
}

// Returns the number of items in the cache.
func (c *SimpleCache) Len() int {
	return len(c.GetALL())
}

// Completely clear the cache
func (c *SimpleCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.init()
}

type simpleItem struct {
	value      interface{}
	expiration *time.Time
}

// returns boolean value whether this item is expired or not.
func (si *simpleItem) IsExpired(now *time.Time) bool {
	if si.expiration == nil {
		return false
	}
	if now == nil {
		t := time.Now()
		now = &t
	}
	return si.expiration.Before(*now)
}
