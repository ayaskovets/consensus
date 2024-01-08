package test

import (
	"math/rand"
	"time"
)

// Customizable Raft settings
type RaftSettings struct {
}

func (settings *RaftSettings) ElectionTimeout() time.Duration {
	min := 150
	max := 300
	return time.Millisecond * time.Duration(rand.Intn(max-min)+min)
}

func (settings *RaftSettings) HeartbeatTimeout() time.Duration {
	min := 50
	max := 100
	return time.Millisecond * time.Duration(rand.Intn(max-min)+min)
}

func (settings *RaftSettings) Majority(peers int) int {
	return peers/2 + 1
}
