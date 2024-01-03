package net

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
)

// Wrapper for RPC client
type Client struct {
	addr string

	mu  sync.Mutex
	rpc *rpc.Client
}

// Construct new client object
func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
		rpc:  nil,
	}
}

// Connect to RPC server.
// Blocking
//
// Idempotent. Returns nil if already connected. Each call to this function
// must be followed by a corresponding disconnect.
func (client *Client) Connect() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.rpc != nil {
		return nil
	}

	var err error
	if client.rpc, err = rpc.Dial("tcp", client.addr); err != nil {
		return err
	}

	log.Printf("connected to %s", client.addr)
	return nil
}

// Invoke RPC method
func (client *Client) Call(serviceMethod string, args any, reply any) error {
	if client.rpc == nil {
		return fmt.Errorf("uninitialized connection to %s", client.addr)
	}
	return client.rpc.Call(serviceMethod, args, reply)
}

// Disconnect from RPC server
// Idempotent. Returns nil if already disconnected
func (client *Client) Disconnect() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.rpc == nil {
		return nil
	}

	if err := client.rpc.Close(); err != nil {
		return err
	}
	client.rpc = nil

	log.Printf("disconnected from %s", client.addr)
	return nil
}
