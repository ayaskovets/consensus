package test

import (
	"net"
	"strings"

	"github.com/ayaskovets/consensus/pkg/consensus/raft"
	"github.com/ayaskovets/consensus/pkg/node"
)

// Remote Raft node implementation
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
