package test

import (
	"testing"

	"github.com/ayaskovets/consensus/pkg/node"
)

func TestNodesConnectivity(t *testing.T) {
	cluster := Cluster{Nodes: []*node.Node{
		node.NewNode(":10001"),
		node.NewNode(":10002"),
		node.NewNode(":10003"),
	}}

	if err := cluster.Up(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cluster.Down(); err != nil {
		t.Fatal(err)
	}
}
