package net

import (
	"log"
	"net/rpc"
	"sync"
)

type Client struct {
	address string

	mu  sync.Mutex
	rpc *rpc.Client
}

func NewClient(address string) *Client {
	return &Client{
		address: address,
		rpc:     nil,
	}
}

func (client *Client) Dial() error {
	client.mu.Lock()

	var err error
	client.rpc, err = rpc.Dial("tcp", client.address)
	if err != nil {
		log.Fatal(err)
	}

	client.mu.Unlock()

	log.Printf("connected to %s", client.address)

	return nil
}
