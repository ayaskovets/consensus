package raft

import (
	"log"
	"sync/atomic"
	"time"
)

const (
	Leader    string = "Leader"
	Follower  string = "Follower"
	Candidate string = "Candidate"

	Nobody int = -1
)

type RaftPeer interface {
	RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error
	AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error
	String() string
}

type RaftNode interface {
	Peers() []RaftPeer
}

type RaftSettings interface {
	HeartbeatTimeout() time.Duration
	ElectionTimeout() time.Duration
	Majority(peers int) int
}

type Raft struct {
	node     RaftNode
	settings RaftSettings

	id          int
	currentTerm int
	votedFor    int
	state       string

	heartbeatTimer *time.Ticker
	electionTimer  *time.Ticker

	shutdown chan any
}

func NewRaft(id int, node RaftNode, settings RaftSettings) *Raft {
	raft := Raft{
		node:     node,
		settings: settings,

		id:          id,
		currentTerm: 0,
		votedFor:    Nobody,
		state:       Follower,

		heartbeatTimer: time.NewTicker(settings.HeartbeatTimeout()),
		electionTimer:  time.NewTicker(settings.ElectionTimeout()),

		shutdown: make(chan any),
	}
	close(raft.shutdown)
	return &raft
}

func (raft *Raft) Up() error {
	select {
	case <-raft.shutdown:
		raft.shutdown = make(chan any)
	default:
		return nil
	}

	log.Printf("%d: start up raft", raft.id)

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

func (raft *Raft) Down() error {
	select {
	case <-raft.shutdown:
		return nil
	default:
		break
	}

	log.Printf("%d: shutdown raft", raft.id)

	raft.electionTimer.Stop()
	close(raft.shutdown)
	return nil
}

func (raft *Raft) sendHearbeats() {
	for _, peer := range raft.node.Peers() {
		go func(peer RaftPeer) {
			args := AppendEntriesArgs{
				Term:     raft.currentTerm,
				LeaderId: raft.id,
			}
			var reply AppendEntriesReply

			if err := peer.AppendEntries(args, &reply); err != nil {
				log.Printf("%d: AppendEntries to %s failed: %s", raft.id, peer, err)
				return
			}

			if reply.Term > raft.currentTerm {
				raft.becomeFollower(reply.Term)
			}
		}(peer)
	}
}

func (raft *Raft) startElection() {
	votes := atomic.Int32{}
	votes.Store(1)
	majority := raft.settings.Majority(len(raft.node.Peers()))
	for _, peer := range raft.node.Peers() {
		go func(peer RaftPeer) {
			args := RequestVoteArgs{
				Term:        raft.currentTerm,
				CandidateId: raft.id,
			}
			var reply RequestVoteReply

			if err := peer.RequestVote(args, &reply); err != nil {
				log.Printf("%d: RequestVote to %s failed: %s", raft.id, peer, err)
				return
			}

			if raft.state != Candidate {
				return
			}

			if reply.Term > raft.currentTerm {
				log.Printf("%d: got request vote reply from %s with term %d > %d", raft.id, peer, reply.Term, raft.currentTerm)
				raft.becomeFollower(reply.Term)
				return
			}

			if reply.VoteGranted == true {
				log.Printf("%d: got vote from %s", raft.id, peer)

				if int(votes.Add(1)) >= majority {
					raft.becomeLeader()
				}
			}
		}(peer)
	}
}

func (raft *Raft) becomeLeader() {
	log.Printf("%d: changed state to Leader with term %d", raft.id, raft.currentTerm)

	raft.state = Leader
	raft.heartbeatTimer.Reset(raft.settings.HeartbeatTimeout())
}

func (raft *Raft) becomeFollower(term int) {
	log.Printf("%d: changed state to Follower with term %d", raft.id, term)

	raft.state = Follower
	raft.currentTerm = term
	raft.votedFor = Nobody
	raft.electionTimer.Reset(raft.settings.ElectionTimeout())
}

func (raft *Raft) becomeCandidate() {
	log.Printf("%d: changed state to Candidate with term %d -> %d", raft.id, raft.currentTerm, raft.currentTerm+1)

	raft.state = Candidate
	raft.currentTerm += 1
	raft.votedFor = raft.id
	raft.startElection()
	raft.electionTimer.Reset(raft.settings.ElectionTimeout())
}
