# metapipe-memcached

A small and performant concurrent memcached client for golang. It uses [protocol pipelining](https://en.wikipedia.org/wiki/Protocol_pipelining) to handle heavy traffic efficiently. 

At the moment, it only supports the latest meta protocol, but adding classic text protocol should be possible.

TODO
----
- docs
- more tests
- timeouts
- TLS
- sharding
- tagged routing
- operation blacklisting
- benchmarking
- monitoring and metrics
- CAS
- classic text protocol
- generics client that takes serializer/deserializer