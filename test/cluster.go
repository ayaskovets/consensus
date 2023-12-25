package test

import (
	"time"

	"github.com/ayaskovets/consensus/pkg/node"
)

// Wrapper for a set of nodes
// All nodes are expected to be hosted on localhost
type Cluster struct {
	nodes []*node.Node
}

// Constructs a cluster from a set of nodes
func NewCluster(nodes []*node.Node) Cluster {
	return Cluster{
		nodes: nodes,
	}
}

// Connects the nodes to each other
// Each node also holds a connection to itself to be able to self-send RPC
// requests
// Blocking
func (cluster *Cluster) Up() error {
	for _, node := range cluster.nodes {
		err := node.Up()
		if err != nil {
			return err
		}
	}
	time.Sleep(time.Millisecond * 100)

	for _, node := range cluster.nodes {
		for _, peer := range cluster.nodes {
			err := node.Connect(peer.Addr())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Shuts down the cluster disconnecting all nodes and stopping RPC servers
func (cluster *Cluster) Down() error {
	for _, node := range cluster.nodes {
		for _, peer := range cluster.nodes {
			err := node.Disconnect(peer.Addr())
			if err != nil {
				return err
			}
		}
	}

	for _, node := range cluster.nodes {
		err := node.Down()
		if err != nil {
			return err
		}
	}

	return nil
}
