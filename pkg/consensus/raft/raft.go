package raft

import (
	"math/rand"
	"time"

	"github.com/ayaskovets/consensus/pkg/node"
)

const (
	Leader    string = "Leader"
	Follower  string = "Follower"
	Candidate string = "Candidate"
)

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

type Raft struct {
	node *node.Node

	id            string
	currentTerm   int
	votedFor      string
	state         string
	electionTimer *time.Ticker

	shutdown chan any
}

func NewRaft(node *node.Node) *Raft {
	raft := Raft{
		node: node,

		id:            node.Addr(),
		currentTerm:   0,
		votedFor:      "",
		state:         Follower,
		electionTimer: time.NewTicker(electionTimeout()),
		shutdown:      make(chan any),
	}

	return &raft
}

func (raft *Raft) Up() error {
	go func() {
		for {
			select {
			case <-raft.shutdown:
				return
			case <-raft.electionTimer.C:
				if raft.state != Leader {
					raft.becomeCandidate()
				}
				break
			default:
				break
			}
		}
	}()

	return nil
}

func (raft *Raft) Down() error {
	raft.electionTimer.Stop()
	close(raft.shutdown)
	return nil
}
