package internal

import (
	"errors"
	"fmt"
)

type BaseClient struct{
	rawClient RawClient
}

func NewBaseClient(addr string, port int, max int) (*BaseClient, error) {
	c, err := NewTcpClient(addr, port, max)
	if err != nil {
		return nil, err
	}
	return &BaseClient{rawClient: c}, nil
}

func (c *BaseClient) Delete(key string) (bool, error) {
	command := fmt.Sprintf("md %s\r\n", key)
	ch := c.rawClient.Dispatch([]byte(command))
	r := <-ch
	if r.Error != nil {
		return false, fmt.Errorf("operation failed: %w", r.Error)
	}
	if len(r.Header) == 0 {
		return false,  errors.New("invalid response")
	}
	switch r.Header[0] {
		case "HD":
			return true, nil
		default:
			return false, nil
    }
}

func (c *BaseClient) Stale(key string) (bool, error) {
	command := fmt.Sprintf("md %s I\r\n", key)
	ch := c.rawClient.Dispatch([]byte(command))
	r := <-ch
	if r.Error != nil {
		return false, fmt.Errorf("operation failed: %w", r.Error)
	}
	if len(r.Header) == 0 {
		return false,  errors.New("invalid response")
	}
	switch r.Header[0] {
		case "HD":
			return true, nil
		default:
			return false, nil
    }
}

func (c *BaseClient) Get(key string) ([]byte, error){
	ch := c.rawClient.Dispatch([]byte(fmt.Sprintf("mg %s t f v\r\n",key)))
	r := <-ch
	return r.Value, nil
}

func (c *BaseClient) Set(key string, value []byte, ttl int) (bool, error) {
	command := fmt.Sprintf("ms %s %d T%d\r\n", key, len(value), ttl)
	dpt := append(append([]byte(command), value...), []byte("\r\n")...)
	ch := c.rawClient.Dispatch(dpt)
	r := <-ch
	if r.Error != nil {
		return false, fmt.Errorf("operation failed: %w", r.Error)
	}
	if len(r.Header) == 0 {
		return false,  errors.New("invalid response")
	}
	switch r.Header[0] {
		case "HD":
			return true, nil
		default:
			return false, nil
    }
}