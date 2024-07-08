package client

type Router interface {
	Route(key string) MemcacheClient
	Shutdown()
}

type DirectRouter struct {
	client MemcacheClient
}

func (r *DirectRouter) Route(key string) MemcacheClient {
	return r.client
}

func (r *DirectRouter) Shutdown() {
	r.client.Shutdown()
}
