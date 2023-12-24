package net

import (
	"log"
	"net/rpc"
)

type Client struct {
	addr string
	rpc  *rpc.Client
}

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

func (client *Client) Dial() error {
	var err error
	client.rpc, err = rpc.Dial("tcp", client.addr)
	if err != nil {
		return err
	}

	log.Printf("connected to %s", client.addr)
	return nil
}

func (client *Client) Close() error {
	err := client.rpc.Close()
	if err != nil {
		return err
	}

	log.Printf("disconnected from %s", client.addr)
	return nil
}
