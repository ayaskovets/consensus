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
	cns := raft.NewRaft(&test.MockRaftNode{}, &test.RaftSettings{})

	assert.Nil(t, cns.Up())
	assert.Nil(t, cns.Down())
}

func TestIdempotency(t *testing.T) {
	cns := raft.NewRaft(&test.MockRaftNode{}, &test.RaftSettings{})

	assert.Nil(t, cns.Up())
	assert.Nil(t, cns.Up())
	assert.Nil(t, cns.Down())
	assert.Nil(t, cns.Down())
}

func TestSingleLeader(t *testing.T) {
	cluster := test.WithCluster(t, addrs(5))

	// Assert that a single leader is elected
	state1 := cluster.WaitElection().GetState()
	assert.NotNil(t, state1.Leader)
}

func TestLeaderDisconnect(t *testing.T) {
	cluster := test.WithCluster(t, addrs(5))

	// Assert that a single leader is elected
	state1 := cluster.WaitElection().GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that after the leader is down a new leader is elected
	state2 := cluster.Disconnect(state1.Leader).WaitElection().GetState()
	assert.NotEqual(t, state2.Leader, state1.Leader)
	assert.Greater(t, state2.Term, state1.Term)
}

func TestNoQuorum(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that a single leader is elected
	state1 := cluster.WaitElection().GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that after the leader is down a new leader is elected
	state2 := cluster.Disconnect(state1.Leader).WaitElection().GetState()
	assert.NotEqual(t, state2.Leader, state1.Leader)
	assert.Greater(t, state2.Term, state1.Term)

	// Assert that for 1/3 alive nodes there is no elected leader
	state3 := cluster.Disconnect(state2.Leader).WaitElection().GetState()
	assert.Nil(t, state3.Leader)
}

func TestNoQuorumReconnect(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	// Assert that a single leader is elected
	state1 := cluster.WaitElection().GetState()
	assert.NotNil(t, state1.Leader)

	// Assert that after the leader is down a new leader is elected
	state2 := cluster.Disconnect(state1.Leader).WaitElection().GetState()
	assert.NotEqual(t, state2.Leader, state1.Leader)
	assert.Greater(t, state2.Term, state1.Term)

	// Assert that for 1/3 alive nodes there is no elected leader
	state3 := cluster.Disconnect(state2.Leader).WaitElection().GetState()
	assert.Nil(t, state3.Leader)

	// Assert that for 2/3 alive nodes after the reconnect a new leader is elected
	state4 := cluster.Reconnect(state2.Leader).WaitElection().GetState()
	assert.NotNil(t, state4.Leader)
	assert.Greater(t, state4.Term, state3.Term)

	// Assert that after all nodes come back there is still a single leader
	state5 := cluster.Reconnect(state1.Leader).WaitElection().GetState()
	assert.NotNil(t, state5.Leader)
	assert.GreaterOrEqual(t, state4.Term, state3.Term)
}
