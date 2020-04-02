package rproxy

import (
	"context"
	"sync"
)

type (
	BatchFunc         func(ctx context.Context, params ...interface{}) (interface{}, error)
	ExtractResultFunc func(resultMap interface{}, param interface{}) interface{}
)

var (
	lock       sync.Mutex
	rpcMapping map[string]BatchFunc
)

func Register(name string, fn BatchFunc) {
	lock.Lock()

	if rpcMapping == nil {
		rpcMapping = make(map[string]BatchFunc, 0)
	}
	rpcMapping[name] = fn

	lock.Unlock()
}
