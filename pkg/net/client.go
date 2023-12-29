package net

import (
	"fmt"
	"log"
	"net/rpc"
)

// Wrapper for an RPC client
type Client struct {
	addr string
	rpc  *rpc.Client
}

// Constructs a new client
// To connect to the RPC server running on the provided address, use Dial()
func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

// Returns the address that is used to connect to the client
func (client *Client) Addr() string {
	return client.addr
}

// Invokes an RPC method on the client
func (client *Client) Call(serviceMethod string, args any, reply any) error {
	if client.rpc == nil {
		return fmt.Errorf("rpc call (%s) on uninitialized connection %s", serviceMethod, client.addr)
	}
	return client.rpc.Call(serviceMethod, args, reply)
}

// Connects to the stored address
// RPC server must be already started on the provided address
// Blocking
func (client *Client) Connect() error {
	var err error
	if client.rpc, err = rpc.Dial("tcp", client.addr); err != nil {
		return err
	}

	log.Printf("connected to %s", client.addr)
	return nil
}

// Closes the connection if it is open
func (client *Client) Disconnect() error {
	if err := client.rpc.Close(); err != nil {
		return err
	}

	log.Printf("disconnected from %s", client.addr)
	return nil
}
