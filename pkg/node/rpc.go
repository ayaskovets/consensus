package node

func (node *Node) Call(addr string, serviceMethod string, args any, reply any) error {
	return node.peers[addr].Call(serviceMethod, args, reply)
}
