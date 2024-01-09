package rpc_test

import (
	"net"
	"net/netip"
	"testing"

	"github.com/ayaskovets/consensus/pkg/rpc"
	"github.com/stretchr/testify/assert"
)

var addr = net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10030"))

type RPC struct{}
type Args struct{}
type Reply struct{}

func (RPC) Call(Args, *Reply) error {
	return nil
}

func TestServerGracefulShutdown(t *testing.T) {
	srv := rpc.NewServer(addr)

	assert.Nil(t, srv.Up())
	assert.Nil(t, srv.Down())
}

func TestServerIdempotency(t *testing.T) {
	srv := rpc.NewServer(addr)

	assert.Nil(t, srv.Up())
	assert.Nil(t, srv.Up())
	assert.Nil(t, srv.Down())
	assert.Nil(t, srv.Down())
}

func TestServerRestart(t *testing.T) {
	srv := rpc.NewServer(addr)

	assert.Nil(t, srv.Up())
	assert.Nil(t, srv.Down())
	assert.Nil(t, srv.Up())
	assert.Nil(t, srv.Down())
}

func TestServerRegister(t *testing.T) {
	srv := rpc.NewServer(addr)

	assert.Nil(t, srv.Register(&RPC{}))
}
