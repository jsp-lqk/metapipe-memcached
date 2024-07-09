package client

import (
	"errors"
	"fmt"
	"sync"
)

var ErrConnectionOverloaded = errors.New("connection overloaded")

// MutationResult contains information about the outcome of a mutation operation (anything but get or info)
type MutationResult int

const (
	Success MutationResult = iota
	Error
	Exists
	NotFound
	NotStored
)

// EntryInfo contains information about a cache entry
type EntryInfo struct {
	TimeToLive  int
	LastAccess  int
	CasId       int
	Fetched     bool
	SlabClassId int
	Size        int
}

// A MemcacheClient is a client implementation that supports memcached operations
type MemcacheClient interface {
	Add(key string, value []byte, ttl int) (MutationResult, error)
	Delete(key string) (MutationResult, error)
	Get(key string) ([]byte, error)
	GetMany(keys []string) (map[string][]byte, error)
	Info(key string) (EntryInfo, error)
	Replace(key string, value []byte, ttl int) (MutationResult, error)
	Set(key string, value []byte, ttl int) (MutationResult, error)
	Touch(key string, ttl int) (MutationResult, error)
	Shutdown()
}

// ConnectionTarget is the information used to locate and connect to a memcached server
type ConnectionTarget struct {
	Address       string
	Port          int
	MaxConcurrent int
}

// A Client is an instance of the metapipe client
// You should be using only this
type Client struct {
	router Router
}

// Creates a Client that connects to a single memcached server
func SingleTargetClient(target ConnectionTarget) (Client, error) {
	ic, err := NewInnerMetaClient(target)
	if err != nil {
		return Client{}, fmt.Errorf("error creating connection: %w", err)
	}
	return Client{
		router: &DirectRouter{client: ic},
	}, nil

}

// Stores an entry ONLY if the key does NOT exist in the server
func (c *Client) Add(key string, value []byte, ttl int) (MutationResult, error) {
	s := c.router.Route(key)
	return s.Add(key, value, ttl)
}

// Deletes an entry
func (c *Client) Delete(key string) (MutationResult, error) {
	s := c.router.Route(key)
	return s.Delete(key)
}

// Gets the contents of an entry
func (c *Client) Get(key string) ([]byte, error) {
	s := c.router.Route(key)
	return s.Get(key)
}

// Gets many entries
func (c *Client) GetMany(keys []string) (map[string][]byte, error) {
	result := make(map[string][]byte, len(keys))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, k := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			r, err := c.Get(key)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				fmt.Printf("error getting key %s: %s", key, err)
				result[key] = nil
			} else {
				result[key] = r
			}
		}(k)
	}
	wg.Wait()
	return result, nil
}

// Gets the information about an entry
func (c *Client) Info(key string) (EntryInfo, error) {
	s := c.router.Route(key)
	return s.Info(key)
}

// Stores an entry ONLY if the key DOES exist in the server
func (c *Client) Replace(key string, value []byte, ttl int) (MutationResult, error) {
	s := c.router.Route(key)
	return s.Replace(key, value, ttl)
}

// Stores an entry
func (c *Client) Set(key string, value []byte, ttl int) (MutationResult, error) {
	s := c.router.Route(key)
	return s.Set(key, value, ttl)
}

// Updates the time to live of an entry
func (c *Client) Touch(key string, ttl int) (MutationResult, error) {
	s := c.router.Route(key)
	return s.Touch(key, ttl)
}

// Shuts down the client that won't accept or return requests anymore
func (c *Client) Shutdown() {
	c.router.Shutdown()
}
