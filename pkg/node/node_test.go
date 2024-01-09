package node_test

import (
	"net"
	"net/netip"
	"testing"

	"github.com/ayaskovets/consensus/pkg/node"
	"github.com/stretchr/testify/assert"
)

var addr1 = net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10010"))
var addr2 = net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10011"))

func TestNodesConnectivity(t *testing.T) {
	node1 := node.NewNode(addr1)
	node2 := node.NewNode(addr2)

	assert.Nil(t, node1.Up())
	assert.Nil(t, node2.Up())
	assert.Nil(t, node1.Connect(addr2))
	assert.Nil(t, node2.Connect(addr1))
	assert.Nil(t, node1.Disconnect(addr2))
	assert.Nil(t, node2.Disconnect(addr1))
	assert.Nil(t, node2.Down())
	assert.Nil(t, node1.Down())
}

func TestNodeIdempotency(t *testing.T) {
	node1 := node.NewNode(addr1)
	node2 := node.NewNode(addr2)

	assert.Nil(t, node1.Up())
	assert.Nil(t, node1.Up())
	assert.Nil(t, node2.Up())
	assert.Nil(t, node2.Up())
	assert.Nil(t, node1.Connect(addr2))
	assert.Nil(t, node1.Connect(addr2))
	assert.Nil(t, node2.Connect(addr1))
	assert.Nil(t, node2.Connect(addr1))
	assert.Nil(t, node1.Disconnect(addr2))
	assert.Nil(t, node1.Disconnect(addr2))
	assert.Nil(t, node2.Disconnect(addr1))
	assert.Nil(t, node2.Disconnect(addr1))
	assert.Nil(t, node2.Down())
	assert.Nil(t, node2.Down())
	assert.Nil(t, node1.Down())
	assert.Nil(t, node1.Down())
}
