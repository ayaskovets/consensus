package test

import (
	"net"
	"testing"

	"github.com/ayaskovets/consensus/pkg/consensus/raft"
	"github.com/ayaskovets/consensus/pkg/node"
)

// Cluster of interconnected localhost nodes
type cluster struct {
	t            *testing.T
	connected    map[net.Addr]RaftNode
	disconnected map[net.Addr]RaftNode
	settings     RaftSettings
}

// Construct new cluster object
func newCluster(addrs []net.Addr) *cluster {
	cluster := cluster{
		t:            nil,
		connected:    make(map[net.Addr]RaftNode),
		disconnected: make(map[net.Addr]RaftNode),
		settings:     NewRaftSettings(),
	}

	for _, addr := range addrs {
		raftnode := RaftNode{node: nil, raft: nil}
		raftnode.node = node.NewNode(addr)
		raftnode.raft = raft.NewRaft(&raftnode, &cluster.settings)
		raftnode.node.Register(raftnode.raft)

		cluster.connected[addr] = raftnode
	}

	return &cluster
}

// Connect nodes to each other and start servers then consensus
//
// Any error is propataged to the corresponding testing.T object
func (cluster *cluster) up() error {
	for _, node := range cluster.connected {
		if err := node.node.Up(); err != nil {
			return err
		}
	}

	for i, node := range cluster.connected {
		for j, peer := range cluster.connected {
			if i == j {
				continue
			}

			if err := node.node.Connect(peer.node.Addr()); err != nil {
				return err
			}
		}
	}

	for _, node := range cluster.connected {
		if err := node.raft.Up(); err != nil {
			return err
		}
	}

	return nil
}

// Disconnect nodes from each other and shutdown consensus then servers
//
// Any error is propataged to the corresponding testing.T object
func (cluster *cluster) down() error {
	for _, node := range cluster.connected {
		if err := node.raft.Down(); err != nil {
			return err
		}
	}

	for _, node := range cluster.disconnected {
		if err := node.raft.Down(); err != nil {
			return err
		}
	}

	for i, node := range cluster.connected {
		for j, peer := range cluster.connected {
			if i == j {
				continue
			}

			if err := node.node.Disconnect(peer.node.Addr()); err != nil {
				return err
			}
		}
	}

	for _, node := range cluster.connected {
		if err := node.node.Down(); err != nil {
			return err
		}
	}

	for _, node := range cluster.disconnected {
		if err := node.node.Down(); err != nil {
			return err
		}
	}

	return nil
}
