package net

import (
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
