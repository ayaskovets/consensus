package raft

import (
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

	go raft.runEventLoop()
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

// RequestVote RPC handler
func (raft *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	raft.mu.Lock()
	defer raft.mu.Unlock()

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

	lastLogIndex, lastLogTerm := raft.logInfo()
	if args.LastLogTerm < lastLogTerm || (args.LastLogTerm == lastLogTerm && args.LastLogIndex < lastLogIndex) {
		reply.Term = raft.currentTerm
		reply.VoteGranted = false
		return nil
	}

	reply.Term = raft.currentTerm
	reply.VoteGranted = true
	raft.votedFor = args.CandidateId
	return nil
}

// AppendEntries RPC handler
func (raft *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	raft.mu.Lock()
	defer raft.mu.Unlock()

	if args.Term < raft.currentTerm {
		reply.Term = raft.currentTerm
		reply.Success = false
		return nil
	}

	if args.Term > raft.currentTerm || raft.state == Candidate {
		raft.becomeFollower(args.Term)
	}

	if args.PrevLogIndex != Initial && (args.PrevLogIndex >= len(raft.log) || args.PrevLogTerm != raft.log[args.PrevLogIndex].Term) {
		reply.Term = raft.currentTerm
		reply.Success = false
		return nil
	}

	logInsertIndex := args.PrevLogIndex + 1
	newEntriesIndex := 0
	for {
		if logInsertIndex >= len(raft.log) || newEntriesIndex >= len(args.Entries) {
			break
		}
		if raft.log[logInsertIndex].Term != args.Entries[newEntriesIndex].Term {
			break
		}
		logInsertIndex++
		newEntriesIndex++
	}

	if newEntriesIndex < len(args.Entries) {
		raft.log = append(raft.log[:logInsertIndex], args.Entries[newEntriesIndex:]...)
	}

	if args.LeaderCommit > raft.commitIndex {
		raft.commitIndex = min(args.LeaderCommit, len(raft.log)-1)
	}

	reply.Term = raft.currentTerm
	reply.Success = true
	raft.electionTimer.Reset(raft.settings.ElectionTimeout())
	return nil
}

// Apply the command to the instance
func (raft *Raft) Apply(command any) bool {
	raft.mu.Lock()
	defer raft.mu.Unlock()

	if raft.state != Leader {
		return false
	}

	raft.log = append(raft.log, Entry{Command: command, Term: raft.currentTerm})
	return true
}

func (raft *Raft) runEventLoop() {
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
}

// Send AppendEntries to all peers
//
// Switch state to Follower in case there is a peer with a greater term
func (raft *Raft) sendHearbeats() {
	currentTerm := raft.currentTerm
	for _, peer := range raft.node.Peers() {
		go func(peer RaftPeer) {
			raft.mu.Lock()
			nextIndex := raft.nextIndex[peer.Id()]
			prevLogIndex := nextIndex - 1
			prevLogTerm := Initial
			if prevLogIndex >= 0 {
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

			if reply.Term > currentTerm {
				raft.becomeFollower(reply.Term)
			}

			if raft.state != Leader {
				return
			}

			if !reply.Success {
				raft.nextIndex[peer.Id()] = nextIndex - 1
				return
			}

			raft.nextIndex[peer.Id()] = nextIndex + len(entries)
			raft.matchIndex[peer.Id()] = raft.nextIndex[peer.Id()] - 1

			commitIndex := raft.commitIndex
			for N := commitIndex + 1; N < len(raft.log); N++ {
				if raft.log[N].Term != currentTerm {
					continue
				}

				matchCount := 1
				for _, peer := range raft.node.Peers() {
					if raft.matchIndex[peer.Id()] < N {
						continue
					}

					matchCount++
				}

				if matchCount >= raft.settings.Majority(len(raft.node.Peers())) {
					raft.commitIndex = N
				}
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
	votes := 1
	majority := raft.settings.Majority(len(raft.node.Peers()))
	currentTerm := raft.currentTerm

	for _, peer := range raft.node.Peers() {
		go func(peer RaftPeer) {
			raft.mu.Lock()
			lastLogIndex, lastLogTerm := raft.logInfo()
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

			if reply.Term > currentTerm {
				raft.becomeFollower(reply.Term)
				return
			}

			if raft.state != Candidate {
				return
			}

			if !reply.VoteGranted {
				return
			}

			if votes++; votes >= majority {
				raft.becomeLeader()
				raft.sendHearbeats()
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
	raft.votedFor = Nobody
	for _, peer := range raft.node.Peers() {
		raft.nextIndex[peer.Id()] = len(raft.log)
		raft.matchIndex[peer.Id()] = Initial
	}
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
	raft.electionTimer.Reset(raft.settings.ElectionTimeout())
}
