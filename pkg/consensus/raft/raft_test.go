package raft_test

import (
	"log"
	"math/rand"
	"net"
	"net/netip"
	"strconv"
	"testing"
	"time"

	"github.com/ayaskovets/consensus/pkg/consensus/raft"
	"github.com/ayaskovets/consensus/pkg/consensus/raft/test"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Millisecond.Nanoseconds())
	log.SetFlags(log.Lmicroseconds)
}

func addrs(n int) []net.Addr {
	addrs := []net.Addr{}
	for i := 0; i < n; i++ {
		addr := "127.0.0.1:" + strconv.Itoa(10000+i)
		addrs = append(addrs, net.TCPAddrFromAddrPort(netip.MustParseAddrPort(addr)))
	}
	return addrs
}

func TestRestart(t *testing.T) {
	cns := raft.NewRaft(&test.MockRaftNode{}, test.NewRaftSettings())

	assert.Nil(t, cns.Up())
	assert.Nil(t, cns.Down())
}

func TestIdempotency(t *testing.T) {
	cns := raft.NewRaft(&test.MockRaftNode{}, test.NewRaftSettings())

	assert.Nil(t, cns.Up())
	assert.Nil(t, cns.Up())
	assert.Nil(t, cns.Down())
	assert.Nil(t, cns.Down())
}

func TestElection(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that a single leader is elected and there is no servers in
	// the Candidate state
	state1 := cluster.
		WaitElection().
		GetState()
	assert.NotNil(t, state1.Leader)
	assert.Len(t, state1.Followers, 2)
	assert.Empty(t, state1.Candidates)
}

func TestFollowerDisconnect(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that the first election is successful
	state1 := cluster.
		WaitElection().
		GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that no election occured after the follower disconnected
	state2 := cluster.
		Disconnect(state1.Followers[0]).
		WaitElection().
		GetState()
	assert.Equal(t, state2.Leader, state1.Leader)
	assert.Equal(t, state2.Term, state1.Term)

	// Assert that after the follower reconnected a new leader was elected
	state3 := cluster.
		Reconnect(state1.Followers[0]).
		WaitElection().
		GetState()
	assert.NotNil(t, state3.Leader)
	assert.Greater(t, state3.Term, state1.Term)
}

func TestLeaderDisconnect(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that the first election is successful
	state1 := cluster.
		WaitElection().
		GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that after the leader disconnected a new leader was elected
	state2 := cluster.
		Disconnect(state1.Leader).
		WaitElection().
		GetState()
	assert.NotEqual(t, state2.Leader, state1.Leader)
	assert.Greater(t, state2.Term, state1.Term)
}

func TestNoQuorum(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that the first election is successful
	state1 := cluster.
		WaitElection().
		GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that after the leader is down a new leader is elected
	state2 := cluster.
		Disconnect(state1.Leader).
		WaitElection().
		GetState()
	assert.NotEqual(t, state2.Leader, state1.Leader)
	assert.Greater(t, state2.Term, state1.Term)

	// Assert that for 1/3 alive nodes there is no elected leader
	state3 := cluster.
		Disconnect(state2.Leader).
		WaitElection().
		GetState()
	assert.Nil(t, state3.Leader)
}

func TestNoQuorumReconnect(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that the first election is successful
	state1 := cluster.
		WaitElection().
		GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that after the leader is down a new leader is elected
	state2 := cluster.
		Disconnect(state1.Leader).
		WaitElection().
		GetState()
	assert.NotEqual(t, state2.Leader, state1.Leader)
	assert.Greater(t, state2.Term, state1.Term)

	// Assert that for 1/3 alive nodes there is no elected leader
	state3 := cluster.
		Disconnect(state2.Leader).
		WaitElection().
		GetState()
	assert.Nil(t, state3.Leader)
	assert.GreaterOrEqual(t, state3.Term, state2.Term)

	// Assert that for 2/3 alive nodes after a new leader is elected
	state4 := cluster.
		Reconnect(state2.Leader).
		WaitElection().
		GetState()
	assert.NotNil(t, state4.Leader)
	assert.Greater(t, state4.Term, state3.Term)

	// Assert that after all nodes come back there is still a single leader
	state5 := cluster.
		Reconnect(state1.Leader).
		WaitElection().
		GetState()
	assert.NotNil(t, state5.Leader)
	assert.GreaterOrEqual(t, state5.Term, state4.Term)
}

func TestApplyOnFollower(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that the first election is successful
	state1 := cluster.
		WaitElection().
		GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that the entry is not committed
	assert.Equal(t, cluster.
		Apply(state1.Followers[0], "one").
		WaitHearbeat().
		CountCommitted([]any{"one"}), 0)
}

func TestApplyOnLeader(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that the first election is successful
	state1 := cluster.
		WaitElection().
		GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that apply requests on leader is successful
	assert.Equal(t, cluster.
		Apply(state1.Leader, "one").
		Apply(state1.Leader, "two").
		WaitHearbeat().
		CountCommitted([]any{"one", "two"}), 3)
}

func TestRepairDisconnectedLog(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that the first election is successful
	state1 := cluster.
		WaitElection().
		GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that apply requests on leader are successful
	assert.Equal(t, cluster.
		Apply(state1.Leader, "one").
		Apply(state1.Leader, "two").
		WaitHearbeat().
		CountCommitted([]any{"one", "two"}), 3)

	// Assert that apply request for 2/3 alive node is successful
	assert.Equal(t, cluster.
		Disconnect(state1.Followers[0]).
		WaitHearbeat().
		Apply(state1.Leader, "three").
		WaitHearbeat().
		CountCommitted([]any{"one", "two", "three"}), 2)

	// Assert that the reconnect is successful with optional re-elections
	state2 := cluster.
		Reconnect(state1.Followers[0]).
		WaitElection().
		GetState()
	assert.NotNil(t, state2.Leader)
	assert.GreaterOrEqual(t, state2.Term, state1.Term)

	// Assert that the disconnected node receives missed updates
	assert.Equal(t, cluster.
		WaitHearbeat().
		CountCommitted([]any{"one", "two", "three"}), 3)
}

func TestApplyNoQuorum(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that the first election is successful
	state1 := cluster.
		WaitElection().
		GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that apply requests on leader are successful
	assert.Equal(t, cluster.
		Apply(state1.Leader, "one").
		Apply(state1.Leader, "two").
		WaitHearbeat().
		CountCommitted([]any{"one", "two"}), 3)

	// Assert that no entries are committed for 1/3 alive nodes
	assert.Equal(t, cluster.
		Disconnect(state1.Followers[0]).
		Disconnect(state1.Followers[1]).
		WaitElection().
		Apply(state1.Leader, "three").
		WaitHearbeat().
		CountCommitted([]any{"one", "two", "three"}), 0)

	// Assert that a new optional elections happed after reconnect
	state2 := cluster.
		Reconnect(state1.Followers[0]).
		Reconnect(state1.Followers[1]).
		WaitElection().
		GetState()
	assert.NotNil(t, state2.Leader)
	assert.GreaterOrEqual(t, state2.Term, state1.Term)

	// Assert that after nodes reconnect new applies are successful
	assert.Equal(t, cluster.
		Apply(state2.Leader, "four").
		WaitHearbeat().
		CountCommitted([]any{"one", "two", "four"}), 3)

}

func TestApplyLeaderDisconnect(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that the first election is successful
	state1 := cluster.WaitElection().GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that apply requests on leader are successful
	assert.Equal(t, cluster.
		Apply(state1.Leader, "one").
		Apply(state1.Leader, "two").
		WaitHearbeat().
		CountCommitted([]any{"one", "two"}), 3)

	// Assert that re-election happens
	state2 := cluster.
		Disconnect(state1.Leader).
		WaitElection().
		GetState()
	assert.NotNil(t, state2.Leader)
	assert.Greater(t, state2.Term, state1.Term)

	// Assert that apply request on the old leader is not committed
	assert.Equal(t, cluster.
		Apply(state1.Leader, "three").
		WaitHearbeat().
		CountCommitted([]any{"one", "two", "three"}), 0)

	// Assert that apply request on the new leader is committed
	assert.Equal(t, cluster.
		Apply(state2.Leader, "four").
		WaitHearbeat().
		CountCommitted([]any{"one", "two", "four"}), 2)

	// Assert that another optional election happens
	state3 := cluster.
		Reconnect(state1.Leader).
		WaitElection().
		GetState()
	assert.NotNil(t, state3.Leader)
	assert.GreaterOrEqual(t, state3.Term, state2.Term)

	// Assert that apply request on the last leader is committed
	assert.Equal(t, cluster.
		Apply(state2.Leader, "five").
		WaitHearbeat().
		CountCommitted([]any{"one", "two", "four", "five"}), 3)
}
