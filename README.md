# metapipe-memcached

A small and performant concurrent memcached client for golang. It uses [protocol pipelining](https://en.wikipedia.org/wiki/Protocol_pipelining) to handle heavy traffic efficiently. 

At the moment, it only supports the latest meta protocol, but adding classic text protocol should be possible.

***WARNING*** This is probably very buggy. I haven't tested it in production yet, apart from integration tests.

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
c, err := client.SingleTargetClient(ConnectionTarget{Address: "127.0.0.1", Port: 11211, MaxConcurrent: 100})
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
