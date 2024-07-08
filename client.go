package client

import (
	"fmt"
	"sync"
)

type MutationResult int

const (
	Success MutationResult = iota
	Error
	Exists
	NotFound
	NotStored
)

type EntryInfo struct {
	TimeToLive  int
	LastAccess  int
	CasId       int
	Fetched     bool
	SlabClassId int
	Size        int
}

type MemcacheClient interface {
	Add(key string, value []byte, ttl int) (MutationResult, error)
	Delete(key string) (MutationResult, error)
	Get(key string) ([]byte, error)
	GetMany(keys []string) (map[string][]byte, error)
	Info(key string) (EntryInfo, error)
	Replace(key string, value []byte, ttl int) (MutationResult, error)
	Set(key string, value []byte, ttl int) (MutationResult, error)
	Stale(key string) (MutationResult, error)
	Touch(key string, ttl int) (MutationResult, error)
	Shutdown()
}

type ConnectionTarget struct {
	Address       string
	Port          int
	MaxConcurrent int
}

type Config struct {
	ConnectionTarget
}

type Client struct {
	router Router
}

func SingleTargetClient(target ConnectionTarget) (Client, error) {
	ic, err := NewInnerMetaClient(target)
	if err != nil {
		return Client{}, fmt.Errorf("error creating connection: %w", err)
	}
	return Client{
		router: &DirectRouter{client: ic},
	}, nil

}

func (c *Client) Add(key string, value []byte, ttl int) (MutationResult, error) {
	s := c.router.Route(key)
	return s.Add(key, value, ttl)
}

func (c *Client) Delete(key string) (MutationResult, error) {
	s := c.router.Route(key)
	return s.Delete(key)
}

func (c *Client) Get(key string) ([]byte, error) {
	s := c.router.Route(key)
	return s.Get(key)
}

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

func (c *Client) Info(key string) (EntryInfo, error) {
	s := c.router.Route(key)
	return s.Info(key)
}

func (c *Client) Replace(key string, value []byte, ttl int) (MutationResult, error) {
	s := c.router.Route(key)
	return s.Replace(key, value, ttl)
}

func (c *Client) Set(key string, value []byte, ttl int) (MutationResult, error) {
	s := c.router.Route(key)
	return s.Set(key, value, ttl)
}

func (c *Client) Stale(key string) (MutationResult, error) {
	s := c.router.Route(key)
	return s.Stale(key)
}

func (c *Client) Touch(key string, ttl int) (MutationResult, error) {
	s := c.router.Route(key)
	return s.Touch(key, ttl)
}

func (c *Client) Shutdown() {
	c.router.Shutdown()
}
