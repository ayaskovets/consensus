package node

import "github.com/ayaskovets/consensus/pkg/net"

// Connects the node to a peer with TCP address addr
// Blocks until the connection is established
func (node *Node) Connect(addr string) error {
	if peer := node.peers[addr]; peer != nil {
		return nil
	}

	client := net.NewClient(addr)
	if err := client.Connect(); err != nil {
		return err
	}

	node.peers[addr] = client
	return nil
}

// Disconnects the node from a peer with TCP address addr
// Non-blocking
func (node *Node) Disconnect(addr string) error {
	peer := node.peers[addr]
	if peer == nil {
		return nil
	}

	if err := peer.Disconnect(); err != nil {
		return err
	}

	delete(node.peers, addr)
	return nil
}
