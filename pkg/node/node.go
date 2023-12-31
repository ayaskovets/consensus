package node

import (
	"fmt"

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

// Returns TCP address:port of the node
func (node *Node) Addr() string {
	return node.server.Addr()
}

// Returns TCP []address:port of node's peers
func (node *Node) Peers() []string {
	peers := make([]string, 0, len(node.peers))
	for addr := range node.peers {
		peers = append(peers, addr)
	}
	return peers
}

// Registers the rcvr object as an RPC receiver
// Can be called multiple times
func (node *Node) Register(rcvr any) error {
	return node.server.Register(rcvr)
}

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

// Invokes RPC method on an object with proveded addr
func (node *Node) Call(addr string, serviceMethod string, args any, reply any) error {
	peer := node.peers[addr]
	if peer == nil {
		return fmt.Errorf("not connected to %s", addr)
	}
	return peer.Call(serviceMethod, args, reply)
}

// Starts up the node by starting its RPC server
// Non-blocking
func (node *Node) Up() error {
	return node.server.Up()
}

// Shuts down the node by stopping its RPC server
// All incoming connections to the node are closed
func (node *Node) Down() error {
	return node.server.Down()
}
