package node

import "fmt"

func (node *Node) Register(rcvr any) error {
	return node.server.Register(rcvr)
}

func (node *Node) Call(addr string, serviceMethod string, args any, reply any) error {
	peer := node.peers[addr]
	if peer == nil {
		return fmt.Errorf("rpc call (%s) on unitialized connection %s", serviceMethod, addr)
	}
	return peer.Call(serviceMethod, args, reply)
}
