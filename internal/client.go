package internal

import (
	"bufio"
	"net"
	"github.com/edwingeng/deque/v2"
	"sync"
)

type RawClient interface {
	Dispatch(r []byte) <-chan Response
}

type Client interface {
	Delete(key string) (bool, error)
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl int) (bool, error)
	Stale(key string) (bool, error)
}

type RequestType int

const (
	ARITHMETIC RequestType = iota
	DEBUG
	DELETE
	GET
	NOOP
	SET
)

type Request struct {
	responseChannel chan Response
}

type Response struct {
	Header []string
	Value []byte
	Error error
}

type ClientConnection struct {
	address    string
	port       int
	maxConcurrent int
	conn       net.Conn
	rw   *bufio.ReadWriter
	mu sync.Mutex
	deque      *deque.Deque[Request]
}
