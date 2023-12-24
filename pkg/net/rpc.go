package net

import "fmt"

func (server *Server) Register(rcvr any) error {
	return server.rpc.Register(rcvr)
}

func (client *Client) Call(serviceMethod string, args any, reply any) error {
	if client.rpc == nil {
		return fmt.Errorf("rpc call (%s) on uninitialized connection %s", serviceMethod, client.addr)
	}
	return client.rpc.Call(serviceMethod, args, reply)
}
