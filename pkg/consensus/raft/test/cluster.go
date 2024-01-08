package test

import (
	"net"
	"testing"

	"github.com/ayaskovets/consensus/pkg/consensus/raft"
	"github.com/ayaskovets/consensus/pkg/node"
)

// Cluster of interconnected localhost nodes
type Cluster struct {
	t     *testing.T
	alive map[net.Addr]RaftNode
	dead  map[net.Addr]RaftNode
}

// Construct new cluster object
func NewCluster(addrs []net.Addr) *Cluster {
	cluster := Cluster{
		alive: make(map[net.Addr]RaftNode),
		dead:  make(map[net.Addr]RaftNode),
	}

	for _, addr := range addrs {
		raftnode := RaftNode{node: nil, raft: nil}
		raftsettings := RaftSettings{}

		raftnode.node = node.NewNode(addr)
		raftnode.raft = raft.NewRaft(&raftnode, &raftsettings)
		raftnode.node.Register(raftnode.raft)

		cluster.alive[addr] = raftnode
	}

	return &cluster
}

// Connect nodes to each other and start servers then consensus
//
// Any error is propataged to the corresponding testing.T object
func (cluster *Cluster) Up() error {
	for _, node := range cluster.alive {
		if err := node.node.Up(); err != nil {
			return err
		}
	}

	for i, node := range cluster.alive {
		for j, peer := range cluster.alive {
			if i == j {
				continue
			}

			if err := node.node.Connect(peer.node.Addr()); err != nil {
				return err
			}
		}
	}

	for _, node := range cluster.alive {
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
	for _, node := range cluster.alive {
		if err := node.raft.Down(); err != nil {
			return err
		}
	}

	for i, node := range cluster.alive {
		for j, peer := range cluster.alive {
			if i == j {
				continue
			}

			if err := node.node.Disconnect(peer.node.Addr()); err != nil {
				return err
			}
		}
	}

	for _, node := range cluster.alive {
		if err := node.node.Down(); err != nil {
			return err
		}
	}

	return nil
}
