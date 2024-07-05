package internal

import (
	"errors"
	"fmt"
)

type BaseClient struct{
	readClient RawClient
	mutationClient *TcpRawClient
}

func NewBaseClient(addr string, port int, max int) (*BaseClient, error) {
	ct := ConnectionTarget{address: addr, port:  port, maxConcurrent: max}
	r, err := NewTcpClient(ct)
	if err != nil {
		return nil, err
	}
	m, err := NewTcpClient(ct)
	if err != nil {
		return nil, err
	}
	return &BaseClient{readClient: r, mutationClient: m}, nil
}

func (c *BaseClient) Delete(key string) (bool, error) {
	command := fmt.Sprintf("md %s\r\n", key)
	return c.mutation([]byte(command))
}

func (c *BaseClient) Stale(key string) (bool, error) {
	command := fmt.Sprintf("ms %s 4 I\r\na\r\n", key)
	return c.mutation([]byte(command))
}

func (c *BaseClient) Get(key string) ([]byte, error){
	ch := c.readClient.Dispatch([]byte(fmt.Sprintf("mg %s t f v\r\n",key)))
	r := <-ch
	if r.Error != nil {
		fmt.Printf("error getting: %s", r.Error.Error())
	}
	return r.Value, nil
}

func (c *BaseClient) Set(key string, value []byte, ttl int) (bool, error) {
	command := fmt.Sprintf("ms %s %d T%d\r\n", key, len(value), ttl)
	dpt := append(append([]byte(command), value...), []byte("\r\n")...)
	return c.mutation(dpt)
}

func (c *BaseClient) mutation(command []byte) (bool, error) {
	ch := c.mutationClient.Dispatch(command)
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