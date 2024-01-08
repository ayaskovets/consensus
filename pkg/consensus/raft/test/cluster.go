package test

import (
	"net"
	"testing"

	"github.com/ayaskovets/consensus/pkg/consensus/raft"
	"github.com/ayaskovets/consensus/pkg/node"
)

// Cluster of interconnected localhost nodes
type Cluster struct {
	nodes map[net.Addr]RaftNode
}

// Construct new cluster object
func NewCluster(addrs []net.Addr) *Cluster {
	cluster := Cluster{
		nodes: make(map[net.Addr]RaftNode),
	}

	for _, addr := range addrs {
		raftnode := RaftNode{node: nil, raft: nil}
		raftsettings := RaftSettings{}

		raftnode.node = node.NewNode(addr)
		raftnode.raft = raft.NewRaft(&raftnode, &raftsettings)
		raftnode.node.Register(raftnode.raft)

		cluster.nodes[addr] = raftnode
	}

	return &cluster
}

// Connect nodes to each other and start servers then consensus
//
// Any error is propataged to the corresponding testing.T object
func (cluster *Cluster) Up() error {
	for _, node := range cluster.nodes {
		if err := node.node.Up(); err != nil {
			return err
		}
	}

	for i, node := range cluster.nodes {
		for j, peer := range cluster.nodes {
			if i == j {
				continue
			}

			if err := node.node.Connect(peer.node.Addr()); err != nil {
				return err
			}
		}
	}

	for _, node := range cluster.nodes {
		if err := node.raft.Up(); err != nil {
			return err
		}
	}

	return nil
}

// Disconnect nodes from each other and shutdown consensus then servers
//
// Any error is propataged to the corresponding testing.T object
func (cluster *Cluster) Down() error {
	for _, node := range cluster.nodes {
		if err := node.raft.Down(); err != nil {
			return err
		}
	}

	for i, node := range cluster.nodes {
		for j, peer := range cluster.nodes {
			if i == j {
				continue
			}

			if err := node.node.Disconnect(peer.node.Addr()); err != nil {
				return err
			}
		}
	}

	for _, node := range cluster.nodes {
		if err := node.node.Down(); err != nil {
			return err
		}
	}

	return nil
}

// Start the provided cluster and auto-cleanup at the end of the test
func WithCluster(t *testing.T, cluster *Cluster) *Cluster {
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

func GetSingleLeader(t *testing.T, cluster *Cluster) net.Addr {
	leaders := []net.Addr{}

	for _, node := range cluster.nodes {
		state, _ := node.raft.Info()
		if state != raft.Leader {
			continue
		}

		leaders = append(leaders, node.node.Addr())
	}

	switch len(leaders) {
	case 0:
		return nil
	case 1:
		break
	default:
		t.Error("multiple leaders")
	}

	return leaders[0]
}
