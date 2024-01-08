package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync/atomic"
)

// Wrapper for RPC client
type Client struct {
	addr net.Addr
	rpc  atomic.Pointer[rpc.Client]
}

// Construct new client object
func NewClient(addr net.Addr) *Client {
	return &Client{
		addr: addr,
		rpc:  atomic.Pointer[rpc.Client]{},
	}
}

// Connect to RPC server.
// Blocking
//
// Idempotent. Returns nil if already connected. Each call to this function
// must be followed by a corresponding disconnect.
func (client *Client) Connect() error {
	if client.rpc.Load() != nil {
		return nil
	}

	client_rpc, err := rpc.Dial(client.addr.Network(), client.addr.String())
	if err != nil {
		return err
	}
	client.rpc.Store(client_rpc)

	log.Printf("connected to %s", client.addr)
	return nil
}

// Invoke RPC method
func (client *Client) Call(serviceMethod string, args any, reply any) error {
	client_rpc := client.rpc.Load()
	if client_rpc == nil {
		return fmt.Errorf("uninitialized connection to %s", client.addr)
	}
	return client_rpc.Call(serviceMethod, args, reply)
}

// Disconnect from RPC server.
//
// Idempotent. Returns nil if already disconnected
func (client *Client) Disconnect() error {
	client_rpc := client.rpc.Load()
	if client_rpc == nil {
		return nil
	}

	if err := client_rpc.Close(); err != nil {
		return err
	}
	client.rpc.Store(nil)

	log.Printf("disconnected from %s", client.addr)
	return nil
}
