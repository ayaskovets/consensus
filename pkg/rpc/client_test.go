package rpc_test

import (
	"testing"

	"github.com/ayaskovets/consensus/pkg/rpc"
)

func TestGracefulShutdown(t *testing.T) {
	srv := rpc.NewServer(addr)
	if err := srv.Up(); err != nil {
		t.Error(err)
	}

	cln := rpc.NewClient(addr)
	if err := cln.Connect(); err != nil {
		t.Error(err)
	}

	if err := cln.Disconnect(); err != nil {
		t.Error(err)
	}

	if err := srv.Down(); err != nil {
		t.Error(err)
	}
}

func TestClientIdempotency(t *testing.T) {
	srv := rpc.NewServer(addr)
	if err := srv.Up(); err != nil {
		t.Error(err)
	}

	cln := rpc.NewClient(addr)
	for i := 0; i < 2; i++ {
		if err := cln.Connect(); err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 2; i++ {
		if err := cln.Disconnect(); err != nil {
			t.Error(err)
		}
	}

	if err := srv.Down(); err != nil {
		t.Error(err)
	}
}

func TestClientCall(t *testing.T) {
	srv := rpc.NewServer(addr)
	if err := srv.Register(&RPC{}); err != nil {
		t.Error(err)
	}

	if err := srv.Up(); err != nil {
		t.Error(err)
	}

	cln := rpc.NewClient(addr)
	if err := cln.Connect(); err != nil {
		t.Error(err)
	}

	if err := cln.Call("RPC.Call", struct{}{}, &struct{}{}); err != nil {
		t.Error(err)
	}

	if err := cln.Disconnect(); err != nil {
		t.Error(err)
	}

	if err := srv.Down(); err != nil {
		t.Error(err)
	}
}

func TestClientReconnect(t *testing.T) {
	srv := rpc.NewServer(addr)
	if err := srv.Register(&RPC{}); err != nil {
		t.Error(err)
	}

	if err := srv.Up(); err != nil {
		t.Error(err)
	}

	cln := rpc.NewClient(addr)
	if err := cln.Connect(); err != nil {
		t.Error(err)
	}

	if err := cln.Call("RPC.Call", struct{}{}, &struct{}{}); err != nil {
		t.Error(err)
	}

	if err := cln.Disconnect(); err != nil {
		t.Error(err)
	}

	if err := cln.Call("RPC.Call", struct{}{}, &struct{}{}); err == nil {
		t.Error(err)
	}

	if err := cln.Connect(); err != nil {
		t.Error(err)
	}

	if err := cln.Call("RPC.Call", struct{}{}, &struct{}{}); err != nil {
		t.Error(err)
	}

	if err := cln.Disconnect(); err != nil {
		t.Error(err)
	}

	if err := srv.Down(); err != nil {
		t.Error(err)
	}
}
