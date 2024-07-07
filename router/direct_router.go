package router

import . "github.com/jsp-lqk/metapipe-memcached"

type DirectRouter struct {
	client Client
}

func (r *DirectRouter) Route(key string) Client {
	return r.client
}

func (r *DirectRouter) Shutdown() {
	r.client.Shutdown()
}