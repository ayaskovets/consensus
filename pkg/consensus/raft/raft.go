package raft

import (
	"log"
	"math/rand"
	"sync/atomic"
	"time"
)

const (
	Leader    string = "Leader"
	Follower  string = "Follower"
	Candidate string = "Candidate"
)

type RaftPeer interface {
	RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error
	AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error
	String() string
}

type RaftNode interface {
	Peers() []RaftPeer
}

type Raft struct {
	node RaftNode

	id            int
	currentTerm   int
	votedFor      int
	state         string
	electionTimer *time.Ticker

	shutdown chan any
}

func NewRaft(id int, node RaftNode) *Raft {
	raft := Raft{
		node: node,

		id:            id,
		currentTerm:   0,
		votedFor:      -1,
		state:         Follower,
		electionTimer: time.NewTicker(electionTimeout()),
		shutdown:      make(chan any),
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

	log.Printf("%d: start up node", raft.id)

	go func() {
		for {
			select {
			case <-raft.shutdown:
				raft.electionTimer.Stop()
			case <-raft.electionTimer.C:
				if raft.state != Leader {
					raft.becomeCandidate()
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

	log.Printf("%d: shutdown node", raft.id)

	raft.electionTimer.Stop()
	close(raft.shutdown)
	return nil
}

func electionTimeout() time.Duration {
	min := 150
	max := 300
	return time.Millisecond * time.Duration(rand.Intn(max-min)+min)
}

func heartbeatTimeout() time.Duration {
	min := 140
	max := 160
	return time.Millisecond * time.Duration(rand.Intn(max-min)+min)
}

func (raft *Raft) becomeLeader() {
	log.Printf("%d: changed state to Leader with term %d", raft.id, raft.currentTerm)

	raft.state = Leader

	go func() {
		heartbeatTimer := time.NewTicker(heartbeatTimeout())
		defer heartbeatTimer.Stop()
		for {
			<-heartbeatTimer.C
			for _, peer := range raft.node.Peers() {
				go func(peer RaftPeer) {
					args := AppendEntriesArgs{
						Term:     raft.currentTerm,
						LeaderId: raft.id,
					}
					var reply AppendEntriesReply

					err := peer.AppendEntries(args, &reply)
					if err != nil {
						log.Printf("%d: AppendEntries to %s failed: %s", raft.id, peer, err)
						return
					}

					if reply.Term > raft.currentTerm {
						raft.becomeFollower(reply.Term)
					}
				}(peer)
			}

			if raft.state != Leader {
				return
			}
		}
	}()
}

func (raft *Raft) becomeFollower(term int) {
	log.Printf("%d: changed state to Follower with term %d", raft.id, term)

	raft.state = Follower
	raft.currentTerm = term
	raft.votedFor = -1
	raft.electionTimer.Reset(electionTimeout())
}

func (raft *Raft) becomeCandidate() {
	log.Printf("%d: changed state to Candidate with term %d -> %d", raft.id, raft.currentTerm, raft.currentTerm+1)

	raft.state = Candidate

	raft.currentTerm += 1
	raft.votedFor = raft.id
	votes := atomic.Int32{}
	votes.Store(1)
	majority := len(raft.node.Peers())/2 + 1
	for _, peer := range raft.node.Peers() {
		go func(peer RaftPeer) {
			args := RequestVoteArgs{
				Term:        raft.currentTerm,
				CandidateId: raft.id,
			}
			var reply RequestVoteReply

			err := peer.RequestVote(args, &reply)
			if err != nil {
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

	raft.electionTimer.Reset(electionTimeout())
}
