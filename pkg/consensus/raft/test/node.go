package test

import (
	"strings"

	"github.com/ayaskovets/consensus/pkg/consensus/raft"
	"github.com/ayaskovets/consensus/pkg/node"
)

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
