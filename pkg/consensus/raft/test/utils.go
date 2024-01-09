package test

import (
	"net"
	"testing"
	"time"

	"github.com/ayaskovets/consensus/pkg/consensus/raft"
)

// Start new cluster and auto-cleanup at the end of the test
func WithCluster(t *testing.T, addrs []net.Addr) *cluster {
	cluster := newCluster(addrs)
	cluster.t = t

	t.Cleanup(func() {
		if err := cluster.down(); err != nil {
			t.Error(err)
		}
	})

	if err := cluster.up(); err != nil {
		t.Error(err)
	}

	return cluster
}

// Shut down all RPC communication with the node with the provided addr
// i.e. make it seem like the node is partitioned
func (cluster *cluster) Disconnect(addr net.Addr) *cluster {
	node, ok := cluster.connected[addr]
	if ok == false {
		return cluster
	}

	for _, peer := range cluster.connected {
		if err := node.node.Disconnect(peer.node.Addr()); err != nil {
			cluster.t.Error(err)
		}
	}

	for _, node := range cluster.connected {
		if node.node.Addr() == addr {
			continue
		}

		if err := node.node.Disconnect(addr); err != nil {
			cluster.t.Error(err)
		}
	}

	cluster.disconnected[addr] = node
	delete(cluster.connected, addr)

	return cluster
}

// Resume RPC communication to and from the node with the provided addr
// i.e. make it seem like the node is network-reachable again
func (cluster *cluster) Reconnect(addr net.Addr) *cluster {
	node, ok := cluster.disconnected[addr]
	if ok == false {
		return cluster
	}

	for _, peer := range cluster.connected {
		if err := node.node.Connect(peer.node.Addr()); err != nil {
			cluster.t.Error(err)
		}
	}

	for _, node := range cluster.connected {
		if node.node.Addr() == addr {
			continue
		}

		if err := node.node.Connect(addr); err != nil {
			cluster.t.Error(err)
		}
	}

	cluster.connected[addr] = node
	delete(cluster.disconnected, addr)

	return cluster
}

// Wait for the next election (approximately)
func (cluster *cluster) WaitElection() *cluster {
	time.Sleep(time.Millisecond * 450)

	return cluster
}

type ClusterState struct {
	Leader     net.Addr
	Term       int
	Followers  []net.Addr
	Candidates []net.Addr
}

// Return addr of the single leader of the cluster and its term.
func (cluster *cluster) GetState() ClusterState {
	info := ClusterState{}
	for _, node := range cluster.connected {
		state, term := node.raft.Info()

		if state == raft.Leader {
			if info.Leader != nil {
				cluster.t.Error("multiple leaders")
			}

			info.Leader = node.node.Addr()
			info.Term = term
		}

		if state == raft.Follower {
			info.Followers = append(info.Followers, node.node.Addr())
		}

		if state == raft.Candidate {
			info.Candidates = append(info.Candidates, node.node.Addr())
		}
	}
	return info
}
