package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setup(t *testing.T) (context.Context, testcontainers.Container, string, int) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "memcached:latest",
		ExposedPorts: []string{"11211/tcp"},
		WaitingFor:   wait.ForListeningPort("11211/tcp"),
	}
	memcachedContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	host, err := memcachedContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := memcachedContainer.MappedPort(ctx, "11211/tcp")
	if err != nil {
		t.Fatal(err)
	}

	return ctx, memcachedContainer, host, port.Int()
}

func TestMetaGetsAndSetsCommands(t *testing.T) {
	ctx, memcachedContainer, host, port := setup(t)
	defer memcachedContainer.Terminate(ctx)

	simpleGetsAndSets(t, host, port)
	allOtherOperations(t, host, port)
	triggerMaxConcurrent(t, host, port)
	triggerTimeout(t, host, port)
}

func triggerMaxConcurrent(t *testing.T, host string, port int) {

	c, err := SingleTargetClient(ConnectionTarget{Address: host, Port: port, MaxOutstandingRequests: 5})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Shutdown()

	var wg sync.WaitGroup
	var maxHit atomic.Bool

	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := c.Set(fmt.Sprintf("key-%d", i), []byte(fmt.Sprintf("value-%d", i)), 0)
			if err != nil {
				if errors.Is(err, ErrConnectionOverloaded) {
					maxHit.Store(true)
				}
			}
		}(i)
	}
	wg.Wait()
	assert.True(t, maxHit.Load(), "Expected to hit the max concurrent limit")
}

func triggerTimeout(t *testing.T, host string, port int) {

	c, err := SingleTargetClient(ConnectionTarget{Address: host, Port: port, MaxOutstandingRequests: 1000, TimeoutMs: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Shutdown()

	var wg sync.WaitGroup
	var timeoutHit atomic.Bool

	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := c.Set(fmt.Sprintf("key-%d", i), []byte(fmt.Sprintf("value-%d", i)), 0)
			if err != nil {
				if errors.Is(err, ErrRequestTimeout) {
					timeoutHit.Store(true)
				}
			}
		}(i)
	}
	wg.Wait()
	assert.True(t, timeoutHit.Load(), "Expected to hit the timeout")
}

func simpleGetsAndSets(t *testing.T, host string, port int) {

	c, err := DefaultClient(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Shutdown()

	// get - not found
	gr, err := c.Get("not-exists")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []byte(nil), gr, "Expected nil response")

	// set - success
	mr, err := c.Set("1", []byte("1"), 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Success, mr, "Expected Success response")

	// get - previously set value
	gr, err = c.Get("1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []byte("1"), gr, "Expected []byte of '1'")

	// set many
	for i := 0; i < 50; i++ {
		mr, err = c.Set(fmt.Sprintf("key-%d", i), []byte(fmt.Sprintf("value-%d", i)), 0)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, Success, mr, "Expected Success response")
	}

	// get many
	keys := make([]string, 50)
	for i := 0; i < 50; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
	}
	mp, err := c.GetMany(keys)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range keys {
		assert.Equal(t, []byte(fmt.Sprintf("value-"+strings.TrimPrefix(k, "key-"))), mp[k], "Unexpected response value")
	}
}

func allOtherOperations(t *testing.T, host string, port int) {
	c, err := DefaultClient(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Shutdown()

	// Add - Stores an entry ONLY if the key does NOT exist in the server
	r, err := c.Add("add-1", []byte("add-1-value"), 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Success, r, "Expected successful mutation operation")

	// get - previously set value
	v, err := c.Get("add-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []byte("add-1-value"), v, "Expected []byte of 'add-1-value'")

	r, err = c.Add("add-1", []byte("add-1-value-1"), 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, NotStored, r, "Expected mutation to not be stored")

	// get - previously set value didn't change
	v, err = c.Get("add-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []byte("add-1-value"), v, "Expected []byte of 'add-1-value'")

	// Replace - Stores an entry ONLY if the key DOES exist in the server
	r, err = c.Replace("replace-1", []byte("replace-1-value"), 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, NotStored, r, "Expected mutation to not be stored")

	// get - previous replace wasn't stored
	v, err = c.Get("replace-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []byte(nil), v, "Expected nil value")

	r, err = c.Set("replace-1", []byte("temp"), 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Success, r, "Expected mutation to be stored successfully")

	r, err = c.Replace("replace-1", []byte("replace-1-value"), 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Success, r, "Expected mutation to be stored successfully")

	// get - previous replace wasn't stored
	v, err = c.Get("replace-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []byte("replace-1-value"), v, "Expected []byte of 'replace-1-value'")

	// Delete - Deletes an entry
	r, err = c.Delete("delete-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, NotFound, r, "Expected mutation to fail because the entry is not found")

	r, err = c.Set("delete-1", []byte("temp"), 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Success, r, "Expected mutation to be stored successfully")

	v, err = c.Get("delete-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []byte("temp"), v, "Expected []byte of 'temp'")

	r, err = c.Delete("delete-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Success, r, "Expected mutation to delete entry successfully")

	v, err = c.Get("delete-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []byte(nil), v, "Expected nil value")

	// Touch - Updates the time to live of an entry
	r, err = c.Touch("nan-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, NotFound, r, "Expected mutation to fail because the entry is not found")

	r, err = c.Set("touch-1", []byte("touch-1-value"), 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Success, r, "Expected mutation to be stored successfully")

	i, err := c.Info("touch-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, -1, i.TimeToLive, "Expected TTL to match initial set")

	// unix timestamp style Touch
	currentTime := time.Now()
	timeIn24Hours := currentTime.Add(24 * time.Hour) // 86400 seconds
	unixTimestamp := int(timeIn24Hours.Unix())

	r, err = c.Touch("touch-1", unixTimestamp)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Success, r, "Expected mutation to update TTL successfully")

	i, err = c.Info("touch-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.LessOrEqual(t, 86400, i.TimeToLive, "Expected TTL to match updated TTL set")

	// seconds based Touch
	r, err = c.Touch("touch-1", 1000)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Success, r, "Expected mutation to update TTL successfully")

	i, err = c.Info("touch-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.LessOrEqual(t, 1000, i.TimeToLive, "Expected TTL to match updated TTL set")

	// Info - Gets the information about an entry
	r, err = c.Set("info-1", []byte("info-1-value"), 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Success, r, "Expected mutation to be stored successfully")

	v, err = c.Get("info-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []byte("info-1-value"), v, "Expected []byte of 'info-1-value'")

	i, err = c.Info("info-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, -1, i.TimeToLive, "Expected TTL to match updated TTL set")
	assert.Equal(t, true, i.Fetched, "Expected fetched to be true")
	assert.Equal(t, 77, i.Size, "Expected size to be 77")
	assert.GreaterOrEqual(t, 60, i.LastAccess, "Expected last access to have happened in the last minute")
}
