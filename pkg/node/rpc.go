package node

import "fmt"

// Registers the rcvr object as an RPC receiver
// Can be called multiple times
func (node *Node) Register(rcvr any) error {
	return node.server.Register(rcvr)
}

// Invokes an RPC method on an object with proveded addr
func (node *Node) Call(addr string, serviceMethod string, args any, reply any) error {
	peer := node.peers[addr]
	if peer == nil {
		return fmt.Errorf("rpc call (%s) on unitialized connection %s", serviceMethod, addr)
	}
	return peer.Call(serviceMethod, args, reply)
}
