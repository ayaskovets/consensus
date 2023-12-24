package node

import "github.com/ayaskovets/consensus/pkg/net"

func (node *Node) Connect(addr string) error {
	peer := node.peers[addr]
	if peer != nil {
		return nil
	}

	client := net.NewClient(addr)
	err := client.Dial()
	if err != nil {
		return err
	}

	node.peers[addr] = client
	return nil
}

func (node *Node) Disconnect(addr string) error {
	peer := node.peers[addr]
	if peer == nil {
		return nil
	}

	err := peer.Close()
	if err != nil {
		return err
	}

	delete(node.peers, addr)
	return nil
}
