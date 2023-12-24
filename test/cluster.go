package test

import (
	"time"

	"github.com/ayaskovets/consensus/pkg/node"
)

type Cluster struct {
	addrs []string
	nodes []*node.Node
}

func NewCluster(addrs []string) Cluster {
	return Cluster{
		addrs: addrs,
		nodes: make([]*node.Node, 0, len(addrs)),
	}
}

func (cluster *Cluster) Up() error {
	for _, addr := range cluster.addrs {
		node := node.NewNode(addr)
		cluster.nodes = append(cluster.nodes, node)
		err := node.Up()
		if err != nil {
			return err
		}
	}
	time.Sleep(time.Millisecond * 100)

	for i, node := range cluster.nodes {
		for j, addr := range cluster.addrs {
			if i == j {
				continue
			}

			err := node.Connect(addr)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (cluster *Cluster) Down() error {
	for i, node := range cluster.nodes {
		for j, addr := range cluster.addrs {
			if i == j {
				continue
			}

			err := node.Disconnect(addr)
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
