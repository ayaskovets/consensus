package test

import (
	"github.com/ayaskovets/consensus/pkg/node"
)

// Wrapper for a set of nodes
// All nodes are expected to be hosted on localhost
type Cluster struct {
	Nodes []*node.Node
}

// Starts up the cluster and connects nodes to each other
// Blocking
func (cluster *Cluster) Up() error {
	for _, node := range cluster.Nodes {
		if err := node.Up(); err != nil {
			return err
		}
	}

	for i, node := range cluster.Nodes {
		for j, peer := range cluster.Nodes {
			if i == j {
				continue
			}

			if err := node.Connect(peer.Addr()); err != nil {
				return err
			}
		}
	}

	return nil
}

// Shuts down the cluster disconnecting all nodes and stopping RPC servers
func (cluster *Cluster) Down() error {
	for i, node := range cluster.Nodes {
		for j, peer := range cluster.Nodes {
			if i == j {
				continue
			}

			if err := node.Disconnect(peer.Addr()); err != nil {
				return err
			}
		}
	}

	for _, node := range cluster.Nodes {
		if err := node.Down(); err != nil {
			return err
		}
	}

	return nil
}
