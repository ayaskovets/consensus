package raft_test

import (
	"log"
	"math/rand"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/ayaskovets/consensus/pkg/consensus/raft"
	"github.com/ayaskovets/consensus/pkg/consensus/raft/test"
	"github.com/stretchr/testify/assert"
)

var addrs = []net.Addr{
	net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10001")),
	net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10002")),
	net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10003")),
	net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10004")),
	net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10005")),
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
	cluster := test.WithCluster(t, addrs)

	cluster.WaitElection()
	leader, _ := cluster.GetSingleLeader()
	assert.NotNil(t, leader)
}

func TestLeaderDisconnect(t *testing.T) {
	cluster := test.WithCluster(t, addrs)

	cluster.WaitElection()
	leader, term := cluster.GetSingleLeader()
	assert.NotNil(t, leader)

	cluster.Disconnect(leader)
	cluster.WaitElection()
	newLeader, newTerm := cluster.GetSingleLeader()
	assert.NotEqual(t, newLeader, leader)
	assert.Greater(t, newTerm, term)
}
