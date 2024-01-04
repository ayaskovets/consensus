package rpc_test

import (
	"net"
	"net/netip"
	"testing"

	"github.com/ayaskovets/consensus/pkg/rpc"
)

var addr = net.TCPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:10000"))

type RPC struct{}
type Args struct{}
type Reply struct{}

func (RPC) Call(Args, *Reply) error {
	return nil
}

func TestServerGracefulShutdown(t *testing.T) {
	srv := rpc.NewServer(addr)
	if err := srv.Up(); err != nil {
		t.Error(err)
	}

	if err := srv.Down(); err != nil {
		t.Error(err)
	}
}

func TestServerIdempotency(t *testing.T) {
	srv := rpc.NewServer(addr)
	for i := 0; i < 2; i++ {
		if err := srv.Up(); err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 2; i++ {
		if err := srv.Down(); err != nil {
			t.Error(err)
		}
	}
}

func TestServerRestart(t *testing.T) {
	srv := rpc.NewServer(addr)
	if err := srv.Up(); err != nil {
		t.Error(err)
	}

	if err := srv.Down(); err != nil {
		t.Error(err)
	}

	if err := srv.Up(); err != nil {
		t.Error(err)
	}

	if err := srv.Down(); err != nil {
		t.Error(err)
	}
}

func TestServerRegister(t *testing.T) {
	srv := rpc.NewServer(addr)
	if err := srv.Register(&RPC{}); err != nil {
		t.Error(err)
	}
}
