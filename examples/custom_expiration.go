package main

import (
	"context"
	"fmt"
	"github.com/bluele/gcache"
	"time"
)

func main() {
	gc := gcache.New(10).
		LFU().
		Build()

	gc.SetWithExpire("key", "ok", time.Second*3)

	v, err := gc.Get(context.Background(), "key")
	if err != nil {
		panic(err)
	}
	fmt.Println("value:", v)

	fmt.Println("waiting 3s for value to expire:")
	time.Sleep(time.Second * 3)

	v, err = gc.Get(context.Background(), "key")
	if err != nil {
		panic(err)
	}
	fmt.Println("value:", v)
}
