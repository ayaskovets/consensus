package test

import (
	"testing"

	"github.com/ayaskovets/consensus/pkg/node"
)

func TestNodesConnectivity(t *testing.T) {
	cluster := NewCluster([]*node.Node{
		node.NewNode(":10001"),
		node.NewNode(":10002"),
		node.NewNode(":10003"),
	})

	err := cluster.Up()
	if err != nil {
		t.Fatal(err)
	}

	err = cluster.Down()
	if err != nil {
		t.Fatal(err)
	}
}
