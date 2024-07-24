# metapipe-memcached

A small and performant concurrent memcached client for golang. It uses [protocol pipelining](https://en.wikipedia.org/wiki/Protocol_pipelining) to handle heavy traffic efficiently. It is safe to be used by multiple concurrent goroutines at the same time.

At the moment, it only supports the latest meta protocol, but adding classic text protocol should be possible.

Docs at [https://pkg.go.dev/github.com/jsp-lqk/metapipe-memcached](https://pkg.go.dev/github.com/jsp-lqk/metapipe-memcached)

## Example

Install and import:
```shell
$ go get github.com/jsp-lqk/metapipe-memcached
```

```golang
import "github.com/jsp-lqk/metapipe-memcached"
``` 

Create a client:  
```go
c, err := client.DefaultClient("127.0.0.1:11211")
```

Get and set:
```go
mr, err := c.Set("key", []byte("value"), 0)
if err != nil {
// handle
}
gr, err = c.Get("key")
if err != nil {
// handle
}
```

If you want to customize connection settings, then instead of `DefaultClient`, use either `SingleTargetClient` (for a single server connection), or `ShardedClient` (for multiple servers). Both take either a single or an array of `ConnectionTarget`, where you can set custom maximum outstanding requests (the amount of requests waiting for response from memcached, default 1000) or the request timeout (in ms, default 1000).

## TODO
- backoff retry
- TLS
- tagged routing
- replicated routing (no sharding)
- operation blacklisting
- benchmarking
- monitoring and metrics
- CAS
- classic text protocol
- generics client that takes serializer/deserializer
