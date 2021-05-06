package main

import (
	"fmt"
	"github.com/bluele/gcache"
	"time"
)

func main () {
	gc := gcache.New(10).
		LRU().
		EvictedFunc(func(key, value interface{}) {
			fmt.Printf("key: [%v] evicted\n", key)
		}).
		ExpireCheckInterval(300 * time.Millisecond).
		Build()

	_ = gc.SetWithExpire(1, 1, time.Second)

	time.Sleep(time.Second * 2)
}
