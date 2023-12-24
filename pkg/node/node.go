package node

import (
	"github.com/ayaskovets/consensus/pkg/net"
)

type Node struct {
	server *net.Server
	peers  map[string]*net.Client
}

func NewNode(addr string) *Node {
	return &Node{
		server: net.NewServer(addr),
		peers:  make(map[string]*net.Client),
	}
}

func (node *Node) Up() error {
	return node.server.Serve()
}

func (node *Node) Down() error {
	for _, peer := range node.peers {
		err := peer.Close()
		if err != nil {
			return err
		}
	}
	return node.server.Close()
}
