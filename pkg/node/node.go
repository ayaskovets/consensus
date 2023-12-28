package node

import (
	"github.com/ayaskovets/consensus/pkg/net"
)

// Consensus-independent vertex in the network of other Node() objects
// that can be possibly located on different remote hosts
// Supports communication with other vertices via RPC and provides network
// topology information
type Node struct {
	server *net.Server
	peers  map[string]*net.Client
}

// Constructs a new node without starting its RPC server
func NewNode(addr string) *Node {
	return &Node{
		server: net.NewServer(addr),
		peers:  make(map[string]*net.Client),
	}
}

// Starts up the node by starting its RPC server
// Non-blocking
func (node *Node) Up() error {
	return node.server.Up()
}

// Shuts down the node by stopping its RPC server
// All incoming connections to the node are closed
func (node *Node) Down() error {
	for _, peer := range node.peers {
		err := peer.Disconnect()
		if err != nil {
			return err
		}
	}
	return node.server.Down()
}
