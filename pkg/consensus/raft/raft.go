package raft

import (
	"log"
	"sync/atomic"
	"time"
)

const (
	// State of Raft instance
	Leader    string = "Leader"
	Follower  string = "Follower"
	Candidate string = "Candidate"

	// Special value for no vote
	Nobody string = ""
)

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

// Local Raft instance state
type Raft struct {
	// Communication with peers and configuration
	node     RaftNode
	settings RaftSettings

	// Raft state
	currentTerm    int
	votedFor       string
	state          string
	heartbeatTimer *time.Ticker
	electionTimer  *time.Ticker

	// Shutdown channel
	shutdown chan any
}

// Construct new Raft object
func NewRaft(node RaftNode, settings RaftSettings) *Raft {
	raft := Raft{
		node:     node,
		settings: settings,

		currentTerm:    0,
		votedFor:       Nobody,
		state:          Follower,
		heartbeatTimer: time.NewTicker(settings.HeartbeatTimeout()),
		electionTimer:  time.NewTicker(settings.ElectionTimeout()),

		shutdown: make(chan any),
	}
	close(raft.shutdown)
	return &raft
}

// Start up local Raft instance.
// Non-blocking
func (raft *Raft) Up() error {
	select {
	case <-raft.shutdown:
		raft.shutdown = make(chan any)
	default:
		return nil
	}

	log.Printf("%s: start up raft", raft.node.Id())

	go func() {
		for {
			select {
			case <-raft.shutdown:
				raft.heartbeatTimer.Stop()
				raft.electionTimer.Stop()
				return
			case <-raft.electionTimer.C:
				if raft.state != Leader {
					raft.becomeCandidate()
				}
			case <-raft.heartbeatTimer.C:
				if raft.state == Leader {
					raft.sendHearbeats()
				}
			default:
			}
		}
	}()

	return nil
}

// Shutdown local Raft instance
func (raft *Raft) Down() error {
	select {
	case <-raft.shutdown:
		return nil
	default:
		break
	}

	log.Printf("%s: shutdown raft", raft.node.Id())

	close(raft.shutdown)
	return nil
}

// RequestVote RPC handler
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

// AppendEntries RPC handler
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

// Send empty AppendEntries to all peers
//
// Switch state to Follower in case there is a peer with a greater term
func (raft *Raft) sendHearbeats() {
	for _, peer := range raft.node.Peers() {
		go func(peer RaftPeer) {
			args := AppendEntriesArgs{
				Term:     raft.currentTerm,
				LeaderId: raft.node.Id(),
			}
			var reply AppendEntriesReply

			if err := peer.AppendEntries(args, &reply); err != nil {
				log.Printf("%s: AppendEntries to %s failed: %s", raft.node.Id(), peer.Id(), err)
				return
			}

			if reply.Term > raft.currentTerm {
				raft.becomeFollower(reply.Term)
			}
		}(peer)
	}
}

// Send RequestVode to all peers
//
// Upon receiving the configured majority of votes, become Leader. If election
// fails with no majority, restart it after the next timeout.
// Switch state to Follower in case there is a peer with a greater term
func (raft *Raft) startElection() {
	votes := atomic.Int32{}
	votes.Store(1)
	majority := raft.settings.Majority(len(raft.node.Peers()))
	for _, peer := range raft.node.Peers() {
		go func(peer RaftPeer) {
			args := RequestVoteArgs{
				Term:        raft.currentTerm,
				CandidateId: raft.node.Id(),
			}
			var reply RequestVoteReply

			if err := peer.RequestVote(args, &reply); err != nil {
				log.Printf("%s: RequestVote to %s failed: %s", raft.node.Id(), peer.Id(), err)
				return
			}

			if raft.state != Candidate {
				return
			}

			if reply.Term > raft.currentTerm {
				raft.becomeFollower(reply.Term)
				return
			}

			if reply.VoteGranted == true {
				log.Printf("%s: received a vote from %s", raft.node.Id(), peer.Id())

				if int(votes.Add(1)) >= majority {
					raft.becomeLeader()
				}
			}
		}(peer)
	}
}

// Switch state to Leader
//
// Enable heartbeats while in Leader state
func (raft *Raft) becomeLeader() {
	log.Printf("%s: Leader in term %d", raft.node.Id(), raft.currentTerm)

	raft.state = Leader
	raft.heartbeatTimer.Reset(raft.settings.HeartbeatTimeout())
}

// Switch state to Follower
//
// Usually Follower state is applied when there is an instance with a greater
// term. The term is provided as an argument
func (raft *Raft) becomeFollower(term int) {
	log.Printf("%s: Follower in term %d", raft.node.Id(), term)

	raft.state = Follower
	raft.currentTerm = term
	raft.votedFor = Nobody
	raft.electionTimer.Reset(raft.settings.ElectionTimeout())
}

// Switch state to Candidate
//
// Triggered upon a heartbeat timeout from the current Leader.
// Increase current term and start a new election voting for self
func (raft *Raft) becomeCandidate() {
	log.Printf("%s: Candidate in term %d", raft.node.Id(), raft.currentTerm+1)

	raft.state = Candidate
	raft.currentTerm += 1
	raft.votedFor = raft.node.Id()
	raft.startElection()
	raft.electionTimer.Reset(raft.settings.ElectionTimeout())
}
