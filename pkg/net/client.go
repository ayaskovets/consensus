package net

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
)

// Wrapper for an RPC client
type Client struct {
	addr string

	mu  sync.Mutex
	rpc *rpc.Client
}

// Constructs a new client
// To connect to the RPC server running on the provided address, use Dial()
func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
		rpc:  nil,
	}
}

// Invokes an RPC method on the client
func (client *Client) Call(serviceMethod string, args any, reply any) error {
	if client.rpc == nil {
		return fmt.Errorf("uninitialized connection to %s", client.addr)
	}
	return client.rpc.Call(serviceMethod, args, reply)
}

// Connects to the stored address.
// RPC server must be already started on the provided address.
// Blocking.
// Idempotent, returns nil if already connected
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

// Closes the connection if it is open.
// Idempotent, returns nil if already disconnected
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
