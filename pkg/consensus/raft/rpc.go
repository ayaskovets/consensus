package raft

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
