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
	if !ok {
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
	if !ok {
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
	time.Sleep(cluster.settings.MaxElectionTimeout * 2)

	return cluster
}

// Wait for the next heartbeat (approximately)
func (cluster *cluster) WaitHearbeat() *cluster {
	time.Sleep(cluster.settings.MaxHeartbeatTimeout * 2)

	return cluster
}

// Apply command to the node with the addr
func (cluster *cluster) Apply(addr net.Addr, command any) *cluster {
	node, ok := cluster.connected[addr]
	if !ok {
		node = cluster.disconnected[addr]
	}

	node.raft.Apply(command)

	return cluster
}

// Return addr of the single leader of the cluster and its term.
type ClusterState struct {
	Leader     net.Addr
	Term       int
	Followers  []net.Addr
	Candidates []net.Addr
}

func (cluster *cluster) GetState() ClusterState {
	info := ClusterState{
		Leader:     nil,
		Term:       raft.Initial,
		Followers:  []net.Addr{},
		Candidates: []net.Addr{},
	}

	for _, node := range cluster.connected {
		nodeInfo := node.raft.GetInfo()

		if nodeInfo.State == raft.Leader {
			if info.Leader != nil {
				cluster.t.Error("multiple leaders")
			}

			info.Leader = node.node.Addr()
			info.Term = nodeInfo.Term
		}

		if nodeInfo.State == raft.Follower {
			info.Followers = append(info.Followers, node.node.Addr())
		}

		if nodeInfo.State == raft.Candidate {
			info.Candidates = append(info.Candidates, node.node.Addr())
		}
	}
	return info
}

// Return number of servers in cluster that have committed entries in the same
// order as in the provided arguments
func (cluster *cluster) CountCommitted(commands []any) int {
	compareLog := func(node RaftNode) bool {
		info := node.raft.GetInfo()

		if info.CommitIndex+1 != len(commands) {
			return false
		}

		for i := 0; i < info.CommitIndex; i++ {
			if info.Log[i].Command != commands[i] {
				return false
			}
		}

		return true
	}

	count := 0
	for _, node := range cluster.connected {
		if compareLog(node) {
			count++
		}
	}
	for _, node := range cluster.disconnected {
		if compareLog(node) {
			count++
		}
	}
	return count
}
