package node

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/ayaskovets/consensus/pkg/rpc"
)

// Consensus-independent node in a peer-to-peer network
type Node struct {
	addr   string
	server *rpc.Server
	peers  map[string]*rpc.Client
}

// Construct a new node object
func NewNode(addr string) *Node {
	return &Node{
		addr:   addr,
		server: rpc.NewServer(net.TCPAddrFromAddrPort(netip.MustParseAddrPort(addr))),
		peers:  make(map[string]*rpc.Client),
	}
}

// Return address of the node
func (node *Node) Addr() string {
	return node.addr
}

// Returns peers addresses
func (node *Node) Peers() []string {
	peers := make([]string, 0, len(node.peers))
	for addr := range node.peers {
		peers = append(peers, addr)
	}
	return peers
}

// Register rcvr object as RPC receiver.
// Having multiple receivers of different types is allowed
func (node *Node) Register(rcvr any) error {
	return node.server.Register(rcvr)
}

// Connect to peer.
// Blocks until the connection is established
//
// Idempotent. Returns nil if already connected
func (node *Node) Connect(addr string) error {
	if peer := node.peers[addr]; peer != nil {
		return nil
	}

	client := rpc.NewClient(net.TCPAddrFromAddrPort(netip.MustParseAddrPort(addr)))
	if err := client.Connect(); err != nil {
		return err
	}

	node.peers[addr] = client
	return nil
}

// Disconnects from peer.
// Non-blocking
//
// Idempotent. Returns nil if already disconnected
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

// Start up node.
// Non-blocking
func (node *Node) Up() error {
	return node.server.Up()
}

// Invoke RPC method on the peer
func (node *Node) Call(addr string, serviceMethod string, args any, reply any) error {
	peer := node.peers[addr]
	if peer == nil {
		return fmt.Errorf("not connected to %s", addr)
	}
	return peer.Call(serviceMethod, args, reply)
}

// Shutdown node
//
// All connections to peers should be closed manually
func (node *Node) Down() error {
	return node.server.Down()
}
