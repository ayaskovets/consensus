package node

import (
	"fmt"
	"net"
	"sync"

	"github.com/ayaskovets/consensus/pkg/rpc"
)

// Consensus-independent node in a peer-to-peer network
type Node struct {
	addr   net.Addr
	server *rpc.Server

	mu    sync.RWMutex
	peers map[net.Addr]*rpc.Client
}

// Construct new node object
func NewNode(addr net.Addr) *Node {
	return &Node{
		addr:   addr,
		server: rpc.NewServer(addr),

		mu:    sync.RWMutex{},
		peers: make(map[net.Addr]*rpc.Client),
	}
}

// Return address of the node
func (node *Node) Addr() net.Addr {
	return node.addr
}

// Returns peers addresses
func (node *Node) Peers() []net.Addr {
	node.mu.RLock()
	defer node.mu.RUnlock()

	peers := make([]net.Addr, 0, len(node.peers))
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
func (node *Node) Connect(addr net.Addr) error {
	node.mu.Lock()
	defer node.mu.Unlock()

	if peer := node.peers[addr]; peer != nil {
		return nil
	}

	client := rpc.NewClient(addr)
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
func (node *Node) Disconnect(addr net.Addr) error {
	node.mu.Lock()
	defer node.mu.Unlock()

	peer := node.peers[addr]
	if peer == nil {
		return nil
	}

	if err := peer.Disconnect(); err != nil {
		return err
	}

	node.peers[addr] = nil
	return nil
}

// Start up node.
// Non-blocking.
//
// Idempotent. Each call to this function must be followed by a
// corresponding shutdown
func (node *Node) Up() error {
	return node.server.Up()
}

// Invoke RPC method on the peer
func (node *Node) Call(addr net.Addr, serviceMethod string, args any, reply any) error {
	node.mu.RLock()
	defer node.mu.RUnlock()

	peer := node.peers[addr]
	if peer == nil {
		return fmt.Errorf("not connected to %s", addr)
	}
	return peer.Call(serviceMethod, args, reply)
}

// Shutdown node
// All outgoing connections to peers of the node should be closed manually
//
// Idempotent. Successful call indicates that node can be started again
func (node *Node) Down() error {
	return node.server.Down()
}
