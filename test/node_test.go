package consensus

import (
	"testing"
	"time"

	"github.com/ayaskovets/consensus/pkg/node"
)

func TestNodesConnectivity(t *testing.T) {
	addrs := []string{":10001", ":10002", ":10003"}

	nodes := make([]*node.Node, 0, len(addrs))
	for _, addr := range addrs {
		nodes = append(nodes, node.NewNode(addr))
	}

	for _, node := range nodes {
		go node.Up()
	}
	time.Sleep(time.Millisecond * 100)

	for _, node := range nodes {
		for _, addr := range addrs {
			err := node.Connect(addr)
			if err != nil {
				t.Fail()
			}
		}
	}

	for _, node := range nodes {
		for _, addr := range addrs {
			err := node.Disconnect(addr)
			if err != nil {
				t.Fail()
			}
		}
	}

	for _, node := range nodes {
		err := node.Down()
		if err != nil {
			t.Fail()
		}
	}
}
