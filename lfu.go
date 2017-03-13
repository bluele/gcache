package gcache

import (
	"container/list"
	"time"
)

// Discards the least frequently used items first.
type LFUCache struct {
	baseCache
	items    map[interface{}]*lfuItem
	freqList *list.List // list for freqEntry
}

func newLFUCache(cb *CacheBuilder) *LFUCache {
	c := &LFUCache{}
	buildCache(&c.baseCache, cb)

	c.init()
	c.loadGroup.cache = c
	return c
}

func (c *LFUCache) init() {
	c.freqList = list.New()
	c.items = make(map[interface{}]*lfuItem, c.size+1)
	c.freqList.PushFront(&freqEntry{
		freq:  0,
		items: make(map[*lfuItem]byte),
	})
}

// set a new key-value pair
func (c *LFUCache) Set(key, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.set(key, value)
	return err
}

func (c *LFUCache) set(key, value interface{}) (interface{}, error) {
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
		item = &lfuItem{
			key:         key,
			value:       value,
			freqElement: nil,
		}
		el := c.freqList.Front()
		fe := el.Value.(*freqEntry)
		fe.items[item] = 1

		item.freqElement = el
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
func (c *LFUCache) Get(key interface{}) (interface{}, error) {
	v, err := c.getValue(key)
	if err == KeyNotFoundError {
		return c.getWithLoader(key, true)
	}
	return v, err
}

// Get a value from cache pool using key if it exists.
// If it dose not exists key, returns KeyNotFoundError.
// And send a request which refresh value for specified key if cache object has LoaderFunc.
func (c *LFUCache) GetIFPresent(key interface{}) (interface{}, error) {
	v, err := c.getValue(key)
	if err == KeyNotFoundError {
		return c.getWithLoader(key, false)
	}
	return v, err
}

func (c *LFUCache) get(key interface{}, onLoad bool) (interface{}, error) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if ok {
		if !item.IsExpired(nil) {
			c.mu.Lock()
			c.increment(item)
			c.mu.Unlock()
			if !onLoad {
				c.stats.IncrHitCount()
			}
			return item, nil
		}
		c.mu.Lock()
		c.removeItem(item)
		c.mu.Unlock()
	}
	if !onLoad {
		c.stats.IncrMissCount()
	}
	return nil, KeyNotFoundError
}

func (c *LFUCache) getValue(key interface{}) (interface{}, error) {
	it, err := c.get(key, false)
	if err != nil {
		return nil, err
	}
	v := it.(*lfuItem).value
	if c.getterFunc != nil {
		return c.getterFunc(key, v)
	}
	return v, nil
}

func (c *LFUCache) getWithLoader(key interface{}, isWait bool) (interface{}, error) {
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
			return v, err
		}
		item := it.(*lfuItem)
		c.increment(item)
		v = item.value
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

func (c *LFUCache) increment(item *lfuItem) {
	currentFreqElement := item.freqElement
	currentFreqEntry := currentFreqElement.Value.(*freqEntry)
	nextFreq := currentFreqEntry.freq + 1
	delete(currentFreqEntry.items, item)

	nextFreqElement := currentFreqElement.Next()
	if nextFreqElement == nil {
		nextFreqElement = c.freqList.InsertAfter(&freqEntry{
			freq:  nextFreq,
			items: make(map[*lfuItem]byte),
		}, currentFreqElement)
	}
	nextFreqElement.Value.(*freqEntry).items[item] = 1
	item.freqElement = nextFreqElement
}

// evict removes the least frequence item from the cache.
func (c *LFUCache) evict(count int) {
	entry := c.freqList.Front()
	for i := 0; i < count; {
		if entry == nil {
			return
		} else {
			for item, _ := range entry.Value.(*freqEntry).items {
				if i >= count {
					return
				}
				c.removeItem(item)
				i++
			}
			entry = entry.Next()
		}
	}
}

// Removes the provided key from the cache.
func (c *LFUCache) Remove(key interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.remove(key)
}

func (c *LFUCache) remove(key interface{}) bool {
	if item, ok := c.items[key]; ok {
		c.removeItem(item)
		return true
	}
	return false
}

// removeElement is used to remove a given list element from the cache
func (c *LFUCache) removeItem(item *lfuItem) {
	delete(c.items, item.key)
	delete(item.freqElement.Value.(*freqEntry).items, item)
	if c.evictedFunc != nil {
		c.evictedFunc(item.key, item.value)
	}
}

func (c *LFUCache) keys() []interface{} {
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
func (c *LFUCache) Keys() []interface{} {
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
func (c *LFUCache) GetALL() map[interface{}]interface{} {
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
func (c *LFUCache) Len() int {
	return len(c.GetALL())
}

// Completely clear the cache
func (c *LFUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.init()
}

type freqEntry struct {
	freq  uint
	items map[*lfuItem]byte
}

type lfuItem struct {
	key         interface{}
	value       interface{}
	freqElement *list.Element
	expiration  *time.Time
}

// returns boolean value whether this item is expired or not.
func (it *lfuItem) IsExpired(now *time.Time) bool {
	if it.expiration == nil {
		return false
	}
	if now == nil {
		t := time.Now()
		now = &t
	}
	return it.expiration.Before(*now)
}
