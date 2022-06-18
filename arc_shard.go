package gcache

import "time"

type ARCShard struct {
	shard []*ARC
	baseCache
}

//var _ ARCShard = *Cache()

var _ Cache = (*ARCShard)(nil)

func newARCShard(cb *CacheBuilder) *ARCShard {
	c := &ARCShard{}
	buildCache(&c.baseCache, cb)
	c.init(cb)
	//c.loadGroup.cache = c

	return c
}

func (c *ARCShard) init(cb *CacheBuilder) {
	c.shard = make([]*ARC, cb.shardCount)
	for i := 0; i < c.shardCount; i++ {
		c.shard[i] = newARC(cb)
	}
}

// GetShard returns shard under given key
func (c *ARCShard) getShard(key interface{}) *ARC {
	return c.shard[uint(fnv32(key))%uint(c.shardCount)]
}

func (c *ARCShard) Set(key, value interface{}) error {
	arc := c.getShard(key)
	_, err := arc.set(key, value)
	return err
}

func (c *ARCShard) SetWithExpire(key, value interface{}, expiration time.Duration) error {
	arc := c.getShard(key)
	return arc.SetWithExpire(key, value, expiration)
}

func (c *ARCShard) Get(key interface{}) (interface{}, error) {
	arc := c.getShard(key)
	v, err := arc.Get(key)
	return v, err
}

func (c *ARCShard) GetIFPresent(key interface{}) (interface{}, error) {
	arc := c.getShard(key)
	return arc.GetIFPresent(key)
}

func (c *ARCShard) GetALL(checkExpired bool) map[interface{}]interface{} {
	//todo
	//return arc.GetALL(checkExpired)
	return map[interface{}]interface{}{}
}

func (c *ARCShard) get(key interface{}, onLoad bool) (interface{}, error) {
	arc := c.getShard(key)
	return arc.get(key, onLoad)
}

func (c *ARCShard) Remove(key interface{}) bool {
	arc := c.getShard(key)
	return arc.Remove(key)
}

func (c *ARCShard) Purge() {
	for _, arc := range c.shard {
		arc.Purge()
	}
	return
}

func (c *ARCShard) Keys(checkExpired bool) []interface{} {
	allKeys := make([]interface{}, 0)
	for _, arc := range c.shard {
		keys := arc.Keys(checkExpired)
		allKeys = append(allKeys, keys)
	}
	return allKeys
}

func (c *ARCShard) Len(checkExpired bool) int {
	len := 0
	for _, arc := range c.shard {
		l := arc.Len(checkExpired)
		len += l
	}
	return len
}

func (c *ARCShard) Has(key interface{}) bool {
	ok := false
	for _, arc := range c.shard {
		if ok = arc.Has(key); ok == true {
			break
		}
	}
	return ok
}
