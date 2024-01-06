package raft

type RequestVoteArgs struct {
	Term        int
	CandidateId int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

func (raft *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	if args.Term > raft.currentTerm {
		raft.becomeFollower(args.Term)
	}

	if args.Term < raft.currentTerm {
		reply.Term = raft.currentTerm
		reply.VoteGranted = false
		return nil
	}

	if raft.votedFor != Nobody && raft.votedFor != args.CandidateId {
		reply.Term = raft.currentTerm
		reply.VoteGranted = false
		return nil
	}

	reply.VoteGranted = true
	raft.votedFor = args.CandidateId

	return nil
}

type AppendEntriesArgs struct {
	Term     int
	LeaderId int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (raft *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	if args.Term > raft.currentTerm {
		raft.becomeFollower(args.Term)
	}

	if args.Term < raft.currentTerm {
		reply.Term = raft.currentTerm
		reply.Success = false
		return nil
	}

	reply.Term = raft.currentTerm
	reply.Success = true
	raft.electionTimer.Reset(raft.settings.ElectionTimeout())

	return nil
}
