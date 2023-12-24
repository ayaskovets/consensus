package net

import (
	"fmt"
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

func (client *Client) Call(serviceMethod string, args any, reply any) error {
	if client.rpc == nil {
		return fmt.Errorf("rpc call (%s) on uninitialized connection %s", serviceMethod, client.addr)
	}
	return client.rpc.Call(serviceMethod, args, reply)
}

func (client *Client) Close() error {
	err := client.rpc.Close()
	if err != nil {
		return err
	}

	log.Printf("disconnected from %s", client.addr)
	return nil
}
