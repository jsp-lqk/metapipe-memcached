package client

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMetaGetsAndSetsCommands(t *testing.T) {
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
	defer memcachedContainer.Terminate(ctx)

	host, err := memcachedContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := memcachedContainer.MappedPort(ctx, "11211/tcp")
	if err != nil {
		t.Fatal(err)
	}

	c, err := SingleTargetClient(ConnectionTarget{Address: host, Port: port.Int(), MaxConcurrent: 100})
	if err != nil {
		t.Fatal(err)
	}

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
	assert.Equal(t, []byte("1"), gr, "Expected byte of '1'")

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

	c.Shutdown()
}
