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

func addrs(n int) []net.Addr {
	addrs := []net.Addr{}
	for i := 0; i < n; i++ {
		addr := "127.0.0.1:" + strconv.Itoa(10000+i)
		addrs = append(addrs, net.TCPAddrFromAddrPort(netip.MustParseAddrPort(addr)))
	}
	return addrs
}

func init() {
	rand.Seed(time.Millisecond.Nanoseconds())
	log.SetFlags(log.Lmicroseconds)
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

	cluster.WaitElection()
	leader, _ := cluster.GetSingleLeader()
	assert.NotNil(t, leader)
}

func TestLeaderDisconnect(t *testing.T) {
	cluster := test.WithCluster(t, addrs(5))

	cluster.WaitElection()
	leader, term := cluster.GetSingleLeader()
	assert.NotNil(t, leader)

	cluster.Disconnect(leader)
	cluster.WaitElection()
	newLeader, newTerm := cluster.GetSingleLeader()
	assert.NotEqual(t, newLeader, leader)
	assert.Greater(t, newTerm, term)
}

func TestNoQuorum(t *testing.T) {
	cluster := test.WithCluster(t, addrs(3))

	cluster.WaitElection()
	leader, term := cluster.GetSingleLeader()
	assert.NotNil(t, leader)

	cluster.Disconnect(leader)
	cluster.WaitElection()
	newLeader, newTerm := cluster.GetSingleLeader()
	assert.NotEqual(t, newLeader, leader)
	assert.Greater(t, newTerm, term)

	cluster.Disconnect(newLeader)
	cluster.WaitElection()
	noLeader, _ := cluster.GetSingleLeader()
	assert.Nil(t, noLeader)
}
