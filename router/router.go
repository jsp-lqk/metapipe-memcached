package router

import (
	. "github.com/jsp-lqk/metapipe-memcached"
)

type Router interface {
	Client(key string) Client
}
