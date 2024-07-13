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

## TODO
- docs
- more tests
- timeouts
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
