package router

import (
	. "github.com/jsp-lqk/metapipe-memcached"
)

type Router interface {
	Route(key string) Client
	Shutdown()
}
