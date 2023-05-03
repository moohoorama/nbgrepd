package nbstore

import (
	"github.com/oxtoacart/bpool"
	"sync"
)

var bufpool *bpool.BytePool
var poolSize int = 1024
var poolUnit int = 1024 * 1024 * 16
var once sync.Once

func SetBytePool(argPoolSize, argPoolUnit int) {
	poolSize = argPoolSize
	poolUnit = argPoolUnit
}

func GetBufpool() *bpool.BytePool {
	once.Do(func() {
		bufpool = bpool.NewBytePool(poolSize, poolUnit)
	})
	return bufpool
}
