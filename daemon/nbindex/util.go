package nbindex

import (
	"github.com/cespare/xxhash/v2"
	_ "github.com/segmentio/fasthash/fnv1a"
)

func Digest(buf []byte) uint32 {
	//return fnv1a.HashBytes32(buf)
	return uint32(xxhash.Sum64(buf))
}
