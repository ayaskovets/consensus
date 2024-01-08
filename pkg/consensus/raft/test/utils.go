package test

import (
	"net"
	"testing"
	"time"

	"github.com/ayaskovets/consensus/pkg/consensus/raft"
)

// Start new cluster and auto-cleanup at the end of the test
func WithCluster(t *testing.T, addrs []net.Addr) *Cluster {
	cluster := NewCluster(addrs)
	cluster.t = t

	t.Cleanup(func() {
		if err := cluster.Down(); err != nil {
			t.Error(err)
		}
	})

	if err := cluster.Up(); err != nil {
		t.Error(err)
	}

	return cluster
}

// Return addr of the single leader of the cluster and its term
func (cluster *Cluster) GetSingleLeader() (net.Addr, int) {
	leaders := []RaftNode{}
	for _, node := range cluster.alive {
		if state, _ := node.raft.Info(); state != raft.Leader {
			continue
		}

		leaders = append(leaders, node)
	}

	switch len(leaders) {
	case 0:
		return nil, 0
	case 1:
		break
	default:
		cluster.t.Error("multiple leaders")
	}

	_, term := leaders[0].raft.Info()
	return leaders[0].node.Addr(), term
}

// Resume RPC communication to and from the node with the provided addr
// i.e. make it seem like the node is back up again
func (cluster *Cluster) Connect(addr net.Addr) {
	node, ok := cluster.dead[addr]
	if ok == true {
		return
	}

	for _, peer := range cluster.alive {
		if err := node.node.Connect(peer.node.Addr()); err != nil {
			cluster.t.Error(err)
		}
	}

	for _, node := range cluster.alive {
		if node.node.Addr() == addr {
			continue
		}

		if err := node.node.Connect(addr); err != nil {
			cluster.t.Error(err)
		}
	}

	cluster.alive[addr] = node
	delete(cluster.dead, addr)
}

// Shut down all RPC communication with the node with the provided addr
// i.e. make it seem like the node is down
func (cluster *Cluster) Disconnect(addr net.Addr) {
	node, ok := cluster.alive[addr]
	if ok == false {
		return
	}

	for _, peer := range cluster.alive {
		if err := node.node.Disconnect(peer.node.Addr()); err != nil {
			cluster.t.Error(err)
		}
	}

	for _, node := range cluster.alive {
		if node.node.Addr() == addr {
			continue
		}

		if err := node.node.Disconnect(addr); err != nil {
			cluster.t.Error(err)
		}
	}

	cluster.dead[addr] = node
	delete(cluster.alive, addr)
}

func (cluster *Cluster) WaitElection() {
	settings := RaftSettings{}
	time.Sleep(settings.ElectionTimeout() * 3)
}
