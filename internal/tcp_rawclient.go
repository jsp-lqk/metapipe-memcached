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
)

type TcpRawClient struct {
	ClientConnection
}

func NewTcpClient(addr string, port int, max int) (*TcpRawClient, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s:%d - %v", addr, port, err)
	}

	clientConn := &TcpRawClient{
		ClientConnection: ClientConnection{
		address: addr,
		port:    port,
		maxConcurrent: max,
		deque: deque.NewDeque[Request](),
		conn:    conn,
		rw: bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		},
	}
	go clientConn.listen()
	return clientConn, nil
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
		tc.deque.PushFront(rq)
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
	}()
	return rc
}

func (tc *TcpRawClient) listen() {
	reader := tc.rw.Reader
	for {
		head, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading from server: %v\n", err)
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
					return
				}
				value = make([]byte, size+2)
				if _, err = io.ReadFull(reader, value); err != nil {
					fmt.Printf("error reading from server: %v\n", err)
					return
				}
			case "CLIENT_ERROR":
				fmt.Printf("error reading from server: %s\n", head)
				panic("error")
		}
		tc.mu.Lock()
		if tc.deque.Len() == 0 {
			panic(fmt.Sprintf("empty deque for response: %s\n", head))
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