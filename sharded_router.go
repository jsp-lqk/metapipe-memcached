package client

import (
	"github.com/dgryski/go-jump"
	"hash/fnv"
)

type ShardedRouter struct {
	clients []MemcacheClient
}

func stringToUint64(s string) uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(s))
	return hasher.Sum64()
}

func (r *ShardedRouter) Route(key string) MemcacheClient {
	i := jump.Hash(stringToUint64(key), len(r.clients))
	return r.clients[i]
}

func (r *ShardedRouter) Shutdown() {
	for _, c := range r.clients {
		c.Shutdown()
	}
}
