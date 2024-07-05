package internal

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/edwingeng/deque/v2"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

type TcpRawClient struct {
	ConnectionTarget
	conn       net.Conn
	mu sync.Mutex
	deque      *deque.Deque[Request]
	rw   *bufio.ReadWriter
}

func NewTcpClient(c ConnectionTarget) (*TcpRawClient, error) {
	tcpRawClient := &TcpRawClient{
		ConnectionTarget: c,
	}
	tcpRawClient.reconnect()
	return tcpRawClient, nil
}

func (tc *TcpRawClient) reconnect() error {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.deque != nil && tc.deque.Len() > 0 {
		for i, n := 0, tc.deque.Len(); i < n; i++ {
			r := tc.deque.PopBack()
			r.responseChannel <- Response{
				Header: nil,
				Value: nil,
				Error: errors.New("connection reset")}
		}
	}
	
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", tc.address, tc.port))
	if err != nil {
		return fmt.Errorf("failed to connect to %s:%d - %v", tc.address, tc.port, err)
	}
	tc.conn = conn
	tc.deque = deque.NewDeque[Request]()
	tc.rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	go tc.listen()
	return nil
}

func (tc *TcpRawClient) Dispatch(r []byte) <-chan Response {
	rc := make(chan Response)
	go func() {
		rq := Request{rc}
		if (tc.deque.Len() > tc.maxConcurrent) {
			rc <- Response{
				Header: nil,
				Value: nil,
				Error: errors.New("connection overloaded")}
			return
		}
		tc.mu.Lock()
		defer tc.mu.Unlock()
		if _, err := tc.rw.Write(r); err != nil {
			rc <- Response{
				Header: nil,
				Value: nil,
				Error: err}
			return
		}
		if err := tc.rw.Flush(); err != nil {
			rc <- Response{
				Header: nil,
				Value: nil,
				Error: err}
			return
		}
		tc.deque.PushFront(rq)
	}()
	return rc
}

func (tc *TcpRawClient) listen() {
	reader := tc.rw.Reader
	for {
		head, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading from server: %v\n", err)
			tc.reconnect()
			return
		}
		var value []byte = nil
		header := strings.Fields(head)
		switch header[0] {
			case "VA":
				sizeString := header[1]
				size, err := strconv.Atoi(sizeString)
				if err != nil {
					fmt.Println("Error converting string to int:", err)
					fmt.Println(head)
					tc.reconnect()
					return
				}
				value = make([]byte, size+2)
				if _, err = io.ReadFull(reader, value); err != nil {
					fmt.Printf("error reading from server: %v\n", err)
					return
				}
			case "CLIENT_ERROR":
				fmt.Printf("error reading from server: %s\n", head)
				tc.reconnect()
				return
		}
		tc.mu.Lock()
		if tc.deque.Len() == 0 {
			tc.mu.Unlock()
			fmt.Printf("empty deque for response: %s\n", head)
			tc.reconnect()
			return
		}
		req := tc.deque.PopBack()
		tc.mu.Unlock()
		req.responseChannel <- Response{
			Header: header,
			Value: value,
			Error: err,
		}
	}
}