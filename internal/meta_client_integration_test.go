package internal

import (
	"context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestMemcached(t *testing.T) {
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

	c, err := NewMetaClient(host, port.Int(), 100)
    if err != nil {
        t.Fatal(err)
    }
    c.Set("1", []byte("1"), 0)

    r, err := c.Get("1")
    if err != nil {
        t.Fatal(err)
    }
    assert.Equal(t, []byte("1"), r, "Expected 'ERROR' response from Memcached")
}
