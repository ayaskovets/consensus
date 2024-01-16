package raft

import "time"

// Remote Raft instance
//
// Implementation is expected to be thread-safe
type RaftPeer interface {
	Id() string
	RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error
	AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error
}

// Customizable network topology info provider for local Raft instance
//
// Implementation is expected to be thread-safe
type RaftNode interface {
	Id() string
	Peers() []RaftPeer
}

// Customizable static-ish settings and configuration for local Raft instance
//
// Implementation is expected to be thread-safe
type RaftSettings interface {
	HeartbeatTimeout() time.Duration
	ElectionTimeout() time.Duration
	Majority(peers int) int
}
