package gcache

import (
	"fmt"
	"reflect"
)

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}
func forEachValue(ifaceSlice interface{}, f func(i int, val interface{})) {
	v := reflect.ValueOf(ifaceSlice)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		panic(fmt.Errorf("forEachValue: expected slice type, found %q", v.Kind().String()))
	}

	for i := 0; i < v.Len(); i++ {
		val := v.Index(i).Interface()
		f(i, val)
	}
}
