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

type BaseTCPClient struct {
	ConnectionTarget
	conn  net.Conn
	mu    sync.Mutex
	deque *deque.Deque[Request]
	rw    *bufio.ReadWriter
}

func NewBaseTCPClient(c ConnectionTarget) (*BaseTCPClient, error) {
	tcpRawClient := &BaseTCPClient{
		ConnectionTarget: c,
	}
	tcpRawClient.reconnect()
	return tcpRawClient, nil
}

func (tc *BaseTCPClient) reconnect() error {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	// on connection loss, clean up the dequeue
	if tc.deque != nil {
		for i, n := 0, tc.deque.Len(); i < n; i++ {
			r := tc.deque.PopBack()
			r.responseChannel <- Response{
				Header: nil,
				Value:  nil,
				Error:  errors.New("connection reset")}
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

func (tc *BaseTCPClient) Dispatch(r []byte) <-chan Response {
	rc := make(chan Response)
	go func() {
		rq := Request{rc}
		if tc.deque.Len() > tc.maxConcurrent {
			rc <- Response{
				Header: nil,
				Value:  nil,
				Error:  errors.New("connection overloaded")}
			return
		}
		tc.mu.Lock()
		defer tc.mu.Unlock()
		if _, err := tc.rw.Write(r); err != nil {
			rc <- Response{
				Header: nil,
				Value:  nil,
				Error:  err}
			return
		}
		if err := tc.rw.Flush(); err != nil {
			rc <- Response{
				Header: nil,
				Value:  nil,
				Error:  err}
			return
		}
		tc.deque.PushFront(rq)
	}()
	return rc
}

func (tc *BaseTCPClient) listen() {
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
		case "VA", "VALUE":
			// only value responses need further reading
			var sizeString string
			if header[0] == "VA" {
				sizeString = header[1]
			} else {
				sizeString = header[3]
			}
			size, err := strconv.Atoi(sizeString)
			if err != nil {
				fmt.Println("fatal connection error parsing response size:", err)
				tc.reconnect()
				return
			}
			value = make([]byte, size+2)
			if _, err = io.ReadFull(reader, value); err != nil {
				fmt.Printf("fatal connection error reading from server: %v\n", err)
				tc.reconnect()
				return
			}
			value = value[:len(value)-2]
		case "ERROR", "CLIENT_ERROR":
			err = fmt.Errorf("error reading from server: %s", header[0])
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
			Value:  value,
			Error:  err,
		}
	}
}
