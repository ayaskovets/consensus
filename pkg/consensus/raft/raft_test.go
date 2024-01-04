package raft_test

import (
	"net"
	"net/netip"
	"testing"

	"github.com/ayaskovets/consensus/pkg/consensus/raft"
	"github.com/stretchr/testify/assert"
)

var addr1 = net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10001"))
var addr2 = net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10002"))
var addr3 = net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10003"))

func TestRestart(t *testing.T) {
	cns := raft.NewRaft(0, nil)
	assert.Nil(t, cns.Up())
	assert.Nil(t, cns.Down())
}

func TestIdempotency(t *testing.T) {
	cns := raft.NewRaft(0, nil)
	assert.Nil(t, cns.Up())
	assert.Nil(t, cns.Up())
	assert.Nil(t, cns.Down())
	assert.Nil(t, cns.Down())
}
