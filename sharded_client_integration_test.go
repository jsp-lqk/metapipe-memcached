package client

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func buildContainer(t *testing.T, port int) (context.Context, testcontainers.Container, string, int) {
	ctx := context.Background()

	portString := fmt.Sprintf("%d/tcp", port)

	req := testcontainers.ContainerRequest{
		Image:        "memcached:latest",
		Entrypoint:   []string{"docker-entrypoint.sh", "-p", fmt.Sprintf("%d", port)},
		ExposedPorts: []string{portString},
		WaitingFor:   wait.ForListeningPort(nat.Port(portString)),
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

	mappedPort, err := memcachedContainer.MappedPort(ctx, nat.Port(portString))
	if err != nil {
		t.Fatal(err)
	}

	return ctx, memcachedContainer, host, mappedPort.Int()
}

func TestShardedGetsAndSets(t *testing.T) {
	targets := make([]ConnectionTarget, 0, 5)
	for i := 0; i <= 4; i++ {
		ctx, c, host, port := buildContainer(t, 11211+i)
		targets = append(targets, ConnectionTarget{Address: host, Port: port, MaxConcurrent: 100})
		defer c.Terminate(ctx)
	}
	shardedTest(t, targets)
}

func shardedTest(t *testing.T, targets []ConnectionTarget) {

	c, err := ShardedClient(targets...)
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
}
