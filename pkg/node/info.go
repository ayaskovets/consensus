package node

// Returns TCP address:port of the node
func (node *Node) Addr() string {
	return node.server.Addr()
}

// Retunrs TCP []address:port of node's peers
func (node *Node) Peers() []string {
	peers := make([]string, 0, len(node.peers))
	for _, peer := range node.peers {
		peers = append(peers, peer.Addr())
	}
	return peers
}
