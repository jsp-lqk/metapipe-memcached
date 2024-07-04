package main

import (
	"fmt"
	"log"
	"testing"
	"runtime"

	"github.com/bradfitz/gomemcache/memcache"
	. "github.com/jsp-lqk/metapipe-memcached/internal"
)

const (
	memcachedServer = "127.0.0.1:11211"
	totalKeys       = 10000
	poolSize        = 50 // Adjust pool size as needed
	parallelism     = 300 // Increase parallelism here
)

func setupMemcache(client *memcache.Client) {
	for i := 0; i < totalKeys; i++ {
		err := client.Set(&memcache.Item{Key: fmt.Sprintf("key%d", i), Value: []byte(fmt.Sprintf("value%d", i))})
		if err != nil {
			log.Fatalf("Failed to set initial data in memcached: %v", err)
		}
	}
}

func BenchmarkMemcacheGet(b *testing.B) {
	client := memcache.New(memcachedServer)

	// Setup initial data
	setupMemcache(client)

	runtime.GOMAXPROCS(parallelism)

	b.ResetTimer()
	

	for i := 0; i < b.N; i++ {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				key := fmt.Sprintf("key%d", i%totalKeys)
				_, err := client.Get(key)
				if err != nil && err != memcache.ErrCacheMiss {
					b.Fatalf("Failed to get key %s: %v", key, err)
				}
			}
		})
	}
}

func BenchmarkMetapipeGet(b *testing.B) {
	client, err := NewBaseClient("127.0.0.1", 11211, 1000)
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	runtime.GOMAXPROCS(parallelism)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				key := fmt.Sprintf("key%d", i%totalKeys)
				_, err := client.Get(key)
				if err != nil {
					b.Fatalf("Failed to get key %s: %v", key, err)
				}
			}
		})
	}
}