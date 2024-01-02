package node

import (
	"testing"

	"github.com/ayaskovets/consensus/pkg/node"
)

func TestNodesConnectivity(t *testing.T) {
	node1 := node.NewNode(":10001")
	if err := node1.Up(); err != nil {
		t.Fatal(err)
	}

	node2 := node.NewNode(":10002")
	if err := node2.Up(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := node1.Connect(node2.Addr()); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := node2.Connect(node1.Addr()); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := node1.Disconnect(node2.Addr()); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := node2.Disconnect(node1.Addr()); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := node2.Down(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := node1.Down(); err != nil {
		t.Fatal(err)
	}
}

func TestNodeIdempotency(t *testing.T) {
	node1 := node.NewNode(":10001")
	if err := node1.Up(); err != nil {
		t.Fatal(err)
	}

	if err := node1.Up(); err != nil {
		t.Fatal(err)
	}

	node2 := node.NewNode(":10002")
	for i := 0; i < 2; i++ {
		if err := node2.Up(); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < 2; i++ {
		if err := node1.Connect(node2.Addr()); err != nil {
			t.Log(err)
			t.Fail()
		}
	}

	for i := 0; i < 2; i++ {
		if err := node2.Connect(node1.Addr()); err != nil {
			t.Log(err)
			t.Fail()
		}
	}

	for i := 0; i < 2; i++ {
		if err := node1.Disconnect(node2.Addr()); err != nil {
			t.Log(err)
			t.Fail()
		}
	}

	for i := 0; i < 2; i++ {
		if err := node2.Disconnect(node1.Addr()); err != nil {
			t.Log(err)
			t.Fail()
		}
	}

	for i := 0; i < 2; i++ {
		if err := node2.Down(); err != nil {
			t.Log(err)
			t.Fail()
		}
	}

	for i := 0; i < 2; i++ {
		if err := node1.Down(); err != nil {
			t.Fatal(err)
		}
	}
}
