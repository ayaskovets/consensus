package test

import (
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/ayaskovets/consensus/pkg/consensus/raft"
	"github.com/ayaskovets/consensus/pkg/node"
)

// Remote Raft node implementation. Can be used as RPC proxy
//
// Uses port number as id. Expects node to be already connected to the addr and running
type RaftPeer struct {
	addr net.Addr
	node *node.Node
}

func (peer *RaftPeer) Id() string {
	port := strings.Split(peer.addr.String(), ":")[1]
	return port
}

func (peer *RaftPeer) RequestVote(args raft.RequestVoteArgs, reply *raft.RequestVoteReply) error {
	return peer.node.Call(peer.addr, "Raft.RequestVote", args, reply)
}

func (peer *RaftPeer) AppendEntries(args raft.AppendEntriesArgs, reply *raft.AppendEntriesReply) error {
	return peer.node.Call(peer.addr, "Raft.AppendEntries", args, reply)
}

// Local Raft node implementation
//
// Expects node to be already up and running. Does not handle startup or
// shutdown
type RaftNode struct {
	node *node.Node
	raft *raft.Raft
}

func (node *RaftNode) Id() string {
	port := strings.Split(node.node.Addr().String(), ":")[1]
	return port
}

func (node *RaftNode) Peers() []raft.RaftPeer {
	peers := make([]raft.RaftPeer, 0)
	for _, addr := range node.node.Peers() {
		peers = append(peers, &RaftPeer{addr: addr, node: node.node})
	}
	return peers
}

// Mock local Raft node to test basic function without starting a real network
// server
type MockRaftNode struct {
}

func (node *MockRaftNode) Id() string {
	return "mockId"
}

func (node *MockRaftNode) Peers() []raft.RaftPeer {
	return []raft.RaftPeer{}
}

// Implementation of raft.RaftSettings
//
// Default timeouts are the ones suggested in the original paper
type RaftSettings struct {
	MinElectionTimeout  time.Duration
	MaxElectionTimeout  time.Duration
	MinHeartbeatTimeout time.Duration
	MaxHeartbeatTimeout time.Duration
}

func NewRaftSettings() RaftSettings {
	return RaftSettings{
		MinElectionTimeout:  150 * time.Millisecond,
		MaxElectionTimeout:  300 * time.Millisecond,
		MinHeartbeatTimeout: 50 * time.Millisecond,
		MaxHeartbeatTimeout: 100 * time.Millisecond,
	}
}

func (settings RaftSettings) ElectionTimeout() time.Duration {
	min := int(settings.MinElectionTimeout.Milliseconds())
	max := int(settings.MaxElectionTimeout.Milliseconds())
	return time.Millisecond * time.Duration(rand.Intn(max-min)+min)
}

func (settings RaftSettings) HeartbeatTimeout() time.Duration {
	min := int(settings.MinHeartbeatTimeout.Milliseconds())
	max := int(settings.MaxHeartbeatTimeout.Milliseconds())
	return time.Millisecond * time.Duration(rand.Intn(max-min)+min)
}

func (settings RaftSettings) Majority(peers int) int {
	return peers/2 + 1
}
