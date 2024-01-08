package test

import "github.com/ayaskovets/consensus/pkg/consensus/raft"

type MockRaftNode struct {
}

func (node *MockRaftNode) Id() string {
	return "mockId"
}

func (node *MockRaftNode) Peers() []raft.RaftPeer {
	return []raft.RaftPeer{}
}
