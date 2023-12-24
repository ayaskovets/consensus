package test

import (
	"time"

	"github.com/ayaskovets/consensus/pkg/node"
)

type Cluster struct {
	nodes []*node.Node
}

func NewCluster(nodes []*node.Node) Cluster {
	return Cluster{
		nodes: nodes,
	}
}

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
