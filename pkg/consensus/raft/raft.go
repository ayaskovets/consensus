package raft

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	// State of Raft instance
	Leader    string = "Leader"
	Follower  string = "Follower"
	Candidate string = "Candidate"

	// Special value for no vote
	Nobody string = ""

	// Special value for uninitialized indices
	Initial int = -1
)

// Local Raft instance state
type Raft struct {
	// Communication with peers configuration
	node     RaftNode
	settings RaftSettings

	// Shutdown channel
	shutdown chan any

	// Event channels
	heartbeatTimer *time.Ticker
	electionTimer  *time.Ticker

	// State mutex
	mu sync.Mutex

	// Persistent state on all servers
	currentTerm int
	votedFor    string
	log         []Entry

	// Volatile state on all servers
	state       string
	commitIndex int
	lastApplied int

	// Volatile Raft state on leaders
	nextIndex  map[string]int
	matchIndex map[string]int
}

// Construct new Raft object
func NewRaft(node RaftNode, settings RaftSettings) *Raft {
	raft := Raft{
		node:     node,
		settings: settings,

		shutdown: make(chan any),

		heartbeatTimer: time.NewTicker(settings.HeartbeatTimeout()),
		electionTimer:  time.NewTicker(settings.ElectionTimeout()),

		mu: sync.Mutex{},

		currentTerm: 0,
		votedFor:    Nobody,
		log:         make([]Entry, 0),

		state:       Follower,
		commitIndex: Initial,
		lastApplied: Initial,

		nextIndex:  make(map[string]int),
		matchIndex: make(map[string]int),
	}
	close(raft.shutdown)
	return &raft
}

// Start up local Raft instance.
// Non-blocking
//
// Idempotent. Returns nil if already started. Each call to this function
// must be followed by a corresponding shutdown
func (raft *Raft) Up() error {
	raft.mu.Lock()
	defer raft.mu.Unlock()

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
				raft.mu.Lock()
				if raft.state != Leader {
					raft.becomeCandidate()
					raft.startElection()
				}
				raft.mu.Unlock()
			case <-raft.heartbeatTimer.C:
				raft.mu.Lock()
				if raft.state == Leader {
					raft.sendHearbeats()
				}
				raft.mu.Unlock()
			default:
			}
		}
	}()

	return nil
}

// Shutdown local Raft instance
//
// Idempotent. Returns nil if already stopped. Successful call indicates that
// the instance is gracefully shut down and can be started again
func (raft *Raft) Down() error {
	raft.mu.Lock()
	defer raft.mu.Unlock()

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

// Apply command to the local instance
//
// Error is returned either if the instance is not Leader at the time or
// in case of any other error
func (raft *Raft) Apply(command any) error {
	raft.mu.Lock()
	defer raft.mu.Unlock()

	if raft.state != Leader {
		return fmt.Errorf("can not apply on %s", raft.state)
	}

	raft.log = append(raft.log, Entry{Command: command, Term: raft.currentTerm})
	return nil
}

// RequestVote RPC handler
func (raft *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	raft.mu.Lock()
	defer raft.mu.Unlock()

	// Convert to Follower on receiving an RPC request greater term (§5.1)
	if args.Term > raft.currentTerm {
		raft.becomeFollower(args.Term)
	}

	reply.Term = raft.currentTerm
	reply.VoteGranted = false

	// Reply false if term < currentTerm (§5.1)
	if args.Term < raft.currentTerm {
		return nil
	}

	// Reply false if votedFor is not null and not equal to candidateId (§5.2)
	if raft.votedFor != Nobody && raft.votedFor != args.CandidateId {
		return nil
	}

	// Reply false if candidate's last log term is not up to date (§5.4)
	lastLogIndex, lastLogTerm := raft.lastLogIndexTerm()
	if args.LastLogTerm < lastLogTerm {
		return nil
	}

	// Reply false if candidate's log has missing entries (§5.4)
	if args.LastLogTerm == lastLogTerm && args.LastLogIndex < lastLogIndex {
		return nil
	}

	// Grant vote otherwise (§5.4)
	reply.VoteGranted = true
	raft.votedFor = args.CandidateId

	return nil
}

// AppendEntries RPC handler
func (raft *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	raft.mu.Lock()
	defer raft.mu.Unlock()

	// Convert to Follower on receiving an RPC request greater term (§5.1)
	if args.Term > raft.currentTerm {
		raft.becomeFollower(args.Term)
	}

	reply.Term = raft.currentTerm
	reply.Success = false

	// Reply false if term < currentTerm (§5.1)
	if args.Term < raft.currentTerm {
		return nil
	}

	// Reply false if log does not contain an entry at PrevLogIndex whose term
	// does matches PrevLogTerm (§5.3)
	if args.PrevLogIndex != Initial {
		if args.PrevLogIndex >= len(raft.log) {
			return nil
		}

		if prevLogTerm := raft.log[args.PrevLogIndex].Term; args.PrevLogTerm != prevLogTerm {
			return nil
		}
	}

	// If an existing entry conflicts with a new one, delete the existing entry
	// and all that follow it (§5.3)
	{
		logIndex := args.PrevLogIndex + 1
		entriesIndex := 0
		for logIndex < len(raft.log) &&
			entriesIndex < len(args.Entries) &&
			raft.log[logIndex].Term == args.Entries[entriesIndex].Term {
			logIndex++
			entriesIndex++
		}

		raft.log = append(raft.log[:logIndex], args.Entries[entriesIndex:]...)
	}

	// Set commitIndex to the Leader commit index
	if args.LeaderCommit > raft.commitIndex {
		raft.commitIndex = min(args.LeaderCommit, len(raft.log)-1)
	}

	// Reply true otherwise
	reply.Success = true
	raft.electionTimer.Reset(raft.settings.ElectionTimeout())

	return nil
}

// Get actual state and current term of the instance
type RaftInfo struct {
	State       string
	Term        int
	Log         []Entry
	CommitIndex int
}

func (raft *Raft) GetInfo() RaftInfo {
	raft.mu.Lock()
	defer raft.mu.Unlock()

	return RaftInfo{
		State:       raft.state,
		Term:        raft.currentTerm,
		Log:         raft.log,
		CommitIndex: raft.commitIndex,
	}
}

// Helper function to get the (LastLogIndex, LastLogTerm) pair
//
// Note that raft.mu has to be locked to get an up-to-date result
func (raft *Raft) lastLogIndexTerm() (int, int) {
	lastLogIndex := Initial
	lastLogTerm := Initial

	if len(raft.log) > 0 {
		lastLogIndex = len(raft.log) - 1
		lastLogTerm = raft.log[lastLogIndex].Term
	}

	return lastLogIndex, lastLogTerm
}

// Send AppendEntries to all peers
//
// Convert to Follower in case there is a peer with a greater term.
// Expects raft.mu to be locked
func (raft *Raft) sendHearbeats() {
	// It is OK to save term globally for all heartbeats because it's change
	// means that a new election occured
	currentTerm := raft.currentTerm

	sendAppendEntries := func(peer RaftPeer) {
		raft.mu.Lock()
		nextIndex := raft.nextIndex[peer.Id()]
		prevLogIndex := nextIndex - 1
		prevLogTerm := Initial
		if prevLogIndex != Initial {
			prevLogTerm = raft.log[prevLogIndex].Term
		}
		entries := raft.log[nextIndex:]
		leaderCommit := raft.commitIndex
		raft.mu.Unlock()

		args := AppendEntriesArgs{
			Term:         currentTerm,
			LeaderId:     raft.node.Id(),
			PrevLogIndex: prevLogIndex,
			PrevLogTerm:  prevLogTerm,
			Entries:      entries,
			LeaderCommit: leaderCommit,
		}
		var reply AppendEntriesReply

		if err := peer.AppendEntries(args, &reply); err != nil {
			return
		}

		raft.mu.Lock()
		defer raft.mu.Unlock()

		// Convert to Follower on receiving an RPC response greater term (§5.1)
		if reply.Term > raft.currentTerm {
			raft.becomeFollower(args.Term)
		}

		// Retry the next heartbeat to this peer with a smaller index
		if !reply.Success {
			raft.nextIndex[peer.Id()] = nextIndex - 1
			return
		}

		raft.nextIndex[peer.Id()] = nextIndex + len(entries)
		raft.matchIndex[peer.Id()] = raft.nextIndex[peer.Id()] - 1

		// If there exists an N such that N > commitIndex, a majority
		// of matchIndex[i] >= N, and log[N].term == currentTerm:
		// set commitIndex = N (§5.3, §5.4).
		commitIndex := raft.commitIndex
		majority := raft.settings.Majority(len(raft.node.Peers()))
		for N := commitIndex + 1; N < len(raft.log); N++ {
			if raft.log[N].Term != currentTerm {
				continue
			}

			matched := 1
			for _, peer := range raft.node.Peers() {
				if raft.matchIndex[peer.Id()] >= N {
					matched++
				}
			}

			if matched >= majority {
				raft.commitIndex = N
			}
		}
	}

	for _, peer := range raft.node.Peers() {
		go sendAppendEntries(peer)
	}
}

// Send RequestVode to all peers
//
// Upon receiving the configured majority of votes, become Leader. If election
// fails with no majority, restart it after the next timeout.
// Convert to Follower in case there is a peer with a greater term.
// Expects raft.mu to be locked
func (raft *Raft) startElection() {
	votes := 1
	majority := raft.settings.Majority(len(raft.node.Peers()))
	currentTerm := raft.currentTerm

	sendRequestVote := func(peer RaftPeer) {
		raft.mu.Lock()
		lastLogIndex, lastLogTerm := raft.lastLogIndexTerm()
		raft.mu.Unlock()

		args := RequestVoteArgs{
			Term:         currentTerm,
			CandidateId:  raft.node.Id(),
			LastLogIndex: lastLogIndex,
			LastLogTerm:  lastLogTerm,
		}
		var reply RequestVoteReply

		if err := peer.RequestVote(args, &reply); err != nil {
			return
		}

		raft.mu.Lock()
		defer raft.mu.Unlock()

		// Convert to Follower on receiving an RPC response greater term (§5.1)
		if reply.Term > raft.currentTerm {
			raft.becomeFollower(args.Term)
		}

		if !reply.VoteGranted {
			return
		}

		// Convert to leader and send initial hearbeats
		if votes++; votes >= majority {
			raft.becomeLeader()
			raft.sendHearbeats()
		}
	}

	for _, peer := range raft.node.Peers() {
		go sendRequestVote(peer)
	}
}

// Convert to Leader
//
// Enable heartbeats while in Leader state. Expects raft.mu to be locked
func (raft *Raft) becomeLeader() {
	log.Printf("%s: Leader in term %d", raft.node.Id(), raft.currentTerm)

	raft.state = Leader
	for _, peer := range raft.node.Peers() {
		raft.nextIndex[peer.Id()] = len(raft.log)
		raft.matchIndex[peer.Id()] = Initial
	}
	raft.heartbeatTimer.Reset(raft.settings.HeartbeatTimeout())
}

// Convert to Follower
//
// Usually Follower state is applied when there is an instance with a greater
// term. The term is provided as an argument. Expects raft.mu to be locked
func (raft *Raft) becomeFollower(term int) {
	log.Printf("%s: Follower in term %d", raft.node.Id(), term)

	raft.state = Follower
	raft.currentTerm = term
	raft.votedFor = Nobody
	raft.electionTimer.Reset(raft.settings.ElectionTimeout())
}

// Convert to Candidate
//
// Triggered upon a heartbeat timeout from the current Leader.
// Increase current term and start a new election voting for self. Expects
// raft.mu to be locked
func (raft *Raft) becomeCandidate() {
	log.Printf("%s: Candidate in term %d", raft.node.Id(), raft.currentTerm+1)

	raft.state = Candidate
	raft.currentTerm++
	raft.votedFor = raft.node.Id()
	raft.electionTimer.Reset(raft.settings.ElectionTimeout())
}
