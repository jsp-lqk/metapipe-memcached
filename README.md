# metapipe-memcached

A small and performant memcached client for golang. It uses the memcached [meta protocol](https://github.com/memcached/memcached/blob/1.6.29/doc/protocol.txt#L79) and [protocol pipelining](https://en.wikipedia.org/wiki/Protocol_pipelining) to handle heavy traffic.

TODO
----
- docs
- tests
- timeouts
- TLS
- sharding
- tagged routing
- operation blacklisting
- benchmarking
- monitoring and metrics
- CAS