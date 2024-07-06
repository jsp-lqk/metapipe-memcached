package internal

import (
	"errors"
	"fmt"
)

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
	Value  []byte
	Error  error
}

type ConnectionTarget struct {
	address       string
	port          int
	maxConcurrent int
}

type MetaClient struct {
	readClient     *BaseTCPClient
	mutationClient *BaseTCPClient
}

func NewMetaClient(addr string, port int, max int) (*MetaClient, error) {
	ct := ConnectionTarget{address: addr, port: port, maxConcurrent: max}
	r, err := NewBaseTCPClient(ct)
	if err != nil {
		return nil, err
	}
	m, err := NewBaseTCPClient(ct)
	if err != nil {
		return nil, err
	}
	return &MetaClient{readClient: r, mutationClient: m}, nil
}

func (c *MetaClient) Delete(key string) (bool, error) {
	command := fmt.Sprintf("md %s\r\n", key)
	return c.mutation([]byte(command))
}

func (c *MetaClient) Stale(key string) (bool, error) {
	command := fmt.Sprintf("md %s I\r\n", key)
	return c.mutation([]byte(command))
}

func (c *MetaClient) Get(key string) ([]byte, error) {
	ch := c.readClient.Dispatch([]byte(fmt.Sprintf("mg %s t f v\r\n", key)))
	r := <-ch
	if r.Error != nil {
		fmt.Printf("error getting: %s", r.Error.Error())
	}
	return r.Value, nil
}

func (c *MetaClient) Set(key string, value []byte, ttl int) (bool, error) {
	command := fmt.Sprintf("ms %s %d T%d\r\n", key, len(value), ttl)
	dpt := append(append([]byte(command), value...), []byte("\r\n")...)
	return c.mutation(dpt)
}

func (c *MetaClient) Touch(key string, ttl int) (bool, error) {
	command := fmt.Sprintf("mg %s T%d\r\n", key, ttl)
	return c.mutation([]byte(command))
}

func (c *MetaClient) mutation(command []byte) (bool, error) {
	ch := c.mutationClient.Dispatch(command)
	r := <-ch
	if r.Error != nil {
		return false, fmt.Errorf("operation failed: %w", r.Error)
	}
	if len(r.Header) == 0 {
		return false, errors.New("invalid response")
	}
	switch r.Header[0] {
	case "HD":
		return true, nil
	default:
		return false, nil
	}
}
