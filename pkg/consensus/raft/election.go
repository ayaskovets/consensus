package raft

import (
	"log"
	"sync/atomic"
	"time"
)

func (raft *Raft) becomeLeader() {
	log.Printf("%s: changed state to Leader with term %d", raft.id, raft.currentTerm)

	raft.state = Leader

	go func() {
		heartbeatTimer := time.NewTicker(heartbeatTimeout())
		defer heartbeatTimer.Stop()
		for {
			<-heartbeatTimer.C
			for _, peer := range raft.node.Peers() {
				go func(peer string) {
					args := AppendEntriesArgs{
						Term:     raft.currentTerm,
						LeaderId: raft.id,
					}
					var reply AppendEntriesReply

					err := raft.node.Call(peer, "Raft.AppendEntries", args, &reply)
					if err != nil {
						log.Printf("%s: AppendEntries to %s failed: %s", raft.id, peer, err)
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
	log.Printf("%s: changed state to Follower with term %d", raft.id, term)

	raft.state = Follower
	raft.currentTerm = term
	raft.votedFor = ""
	raft.electionTimer.Reset(electionTimeout())
}

func (raft *Raft) becomeCandidate() {
	log.Printf("%s: changed state to Candidate with term %d -> %d", raft.id, raft.currentTerm, raft.currentTerm+1)

	raft.state = Candidate

	raft.currentTerm += 1
	raft.votedFor = raft.id
	votes := atomic.Int32{}
	votes.Store(1)
	majority := len(raft.node.Peers())/2 + 1
	for _, peer := range raft.node.Peers() {
		go func(peer string) {
			args := RequestVoteArgs{
				Term:        raft.currentTerm,
				CandidateId: raft.id,
			}
			var reply RequestVoteReply

			err := raft.node.Call(peer, "Raft.RequestVote", args, &reply)
			if err != nil {
				log.Printf("%s: RequestVote to %s failed: %s", raft.id, peer, err)
				return
			}

			if raft.state != Candidate {
				return
			}

			if reply.Term > raft.currentTerm {
				log.Printf("%s: got request vote reply from %s with term %d > %d", raft.id, peer, reply.Term, raft.currentTerm)
				raft.becomeFollower(reply.Term)
				return
			}

			if reply.VoteGranted == true {
				log.Printf("%s: got vote from %s", raft.id, peer)

				if int(votes.Add(1)) >= majority {
					raft.becomeLeader()
				}
			}
		}(peer)
	}

	raft.electionTimer.Reset(electionTimeout())
}
