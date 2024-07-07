package internal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	. "github.com/jsp-lqk/metapipe-memcached"
)

type Request struct {
	responseChannel chan Response
}

type Response struct {
	Header []string
	Value  []byte
	Error  error
}

type ConnectionTarget struct {
	address       string
	port          int
	maxConcurrent int
}

type MetaClient struct {
	readClient     *BaseTCPClient
	mutationClient *BaseTCPClient
}

func NewMetaClient(addr string, port int, max int) (*MetaClient, error) {
	ct := ConnectionTarget{address: addr, port: port, maxConcurrent: max}
	r, err := NewBaseTCPClient(ct)
	if err != nil {
		return nil, err
	}
	m, err := NewBaseTCPClient(ct)
	if err != nil {
		return nil, err
	}
	return &MetaClient{readClient: r, mutationClient: m}, nil
}

func (c *MetaClient) Shutdown() {
	c.mutationClient.Shutdown()
	c.readClient.Shutdown()
}

func (c *MetaClient) Info(key string) (EntryInfo, error) {
	ch := c.readClient.Dispatch([]byte(fmt.Sprintf("me %s\r\n", key)))
	r := <-ch
	if r.Error != nil {
		return EntryInfo{}, fmt.Errorf("operation failed: %w", r.Error)
	}
	switch r.Header[0] {
	case "ME":
		if len(r.Header) < 8 {
			return EntryInfo{}, fmt.Errorf("invalid response size: %d", len(r.Header))
		} 
		return headerToEntryInfo(r.Header)
	default:
		return EntryInfo{}, fmt.Errorf("invalid response: %s", r.Header[0])
	}
}

func headerToEntryInfo(header []string) (EntryInfo, error) {
	ttl, err := strconv.Atoi(getDebugValue(header[2]))
	if err != nil {
		return EntryInfo{}, fmt.Errorf("fatal connection error parsing header: %w", err)
	}

	la, err := strconv.Atoi(getDebugValue(header[3]))
	if err != nil {
		return EntryInfo{}, fmt.Errorf("fatal connection error parsing header: %w", err)
	}

	casId, err := strconv.Atoi(getDebugValue(header[4]))
	if err != nil {
		return EntryInfo{}, fmt.Errorf("fatal connection error parsing header: %w", err)
	}

	var fetch bool
	if getDebugValue(header[5]) == "yes" {
		fetch = true
	}

	cls, err := strconv.Atoi(getDebugValue(header[6]))
	if err != nil {
		return EntryInfo{}, fmt.Errorf("fatal connection error parsing header: %w", err)
	}

	size, err := strconv.Atoi(getDebugValue(header[7]))
	if err != nil {
		return EntryInfo{}, fmt.Errorf("fatal connection error parsing header: %w", err)
	}

	return EntryInfo{
		TimeToLive: ttl,
		LastAccess: la,
		CasId: casId,
		Fetched: fetch,
		SlabClassId: cls,
		Size: size,
	}, nil
}

func getDebugValue(input string) string {
	index := strings.Index(input, "=")
	if index == -1 {
		return ""
	}
	return input[index+1:]
}

func (c *MetaClient) Delete(key string) (MutationResult, error) {
	command := fmt.Sprintf("md %s\r\n", key)
	return c.mutation([]byte(command))
}

func (c *MetaClient) Stale(key string) (MutationResult, error) {
	command := fmt.Sprintf("md %s I\r\n", key)
	return c.mutation([]byte(command))
}

func (c *MetaClient) Get(key string) ([]byte, error) {
	ch := c.readClient.Dispatch([]byte(fmt.Sprintf("mg %s t f v\r\n", key)))
	r := <-ch
	if r.Error != nil {
		return nil, fmt.Errorf("operation failed: %w", r.Error)
	}
	switch r.Header[0] {
	case "VA":
		return r.Value, nil
	case "EN":
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid response: %s", r.Header[0])
	}
}

func (c *MetaClient) GetMany(keys []string) (map[string][]byte, error) {
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

func (c *MetaClient) Set(key string, value []byte, ttl int) (MutationResult, error) {
	command := fmt.Sprintf("ms %s %d T%d\r\n", key, len(value), ttl)
	dpt := append(append([]byte(command), value...), []byte("\r\n")...)
	return c.mutation(dpt)
}

func (c *MetaClient) Touch(key string, ttl int) (MutationResult, error) {
	command := fmt.Sprintf("mg %s T%d\r\n", key, ttl)
	return c.mutation([]byte(command))
}

func (c *MetaClient) mutation(command []byte) (MutationResult, error) {
	ch := c.mutationClient.Dispatch(command)
	r := <-ch
	if r.Error != nil {
		return Error, fmt.Errorf("operation failed: %w", r.Error)
	}
	if len(r.Header) == 0 {
		return Error, errors.New("empty response")
	}
	switch r.Header[0] {
	case "HD":
		return Success, nil
	case "NS":
		return NotStored, nil
	case "EX":
		return Exists, nil
	case "NF", "EN":
		return NotFound, nil
	default:
		return Error, fmt.Errorf("invalid response: %s", r.Header[0])
	}
}
