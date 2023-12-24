package node

func (node *Node) Addr() string {
	return node.server.Addr()
}
