package raft

import "time"

// RequestVote RPC args
type RequestVoteArgs struct {
	Term        int
	CandidateId string
}

// RequestVote RPC reply
type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

// AppendEntries RPC args
type AppendEntriesArgs struct {
	Term     int
	LeaderId string
}

// AppendEntries RPC reply
type AppendEntriesReply struct {
	Term    int
	Success bool
}

// Remote Raft instance interface
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
