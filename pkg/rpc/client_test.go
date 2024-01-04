package rpc_test

import (
	"testing"

	"github.com/ayaskovets/consensus/pkg/rpc"
	"github.com/stretchr/testify/assert"
)

func TestGracefulShutdown(t *testing.T) {
	srv := rpc.NewServer(addr)
	cln := rpc.NewClient(addr)

	assert.Nil(t, srv.Up())
	assert.Nil(t, cln.Connect())
	assert.Nil(t, cln.Disconnect())
	assert.Nil(t, srv.Down())
}

func TestClientIdempotency(t *testing.T) {
	srv := rpc.NewServer(addr)
	cln := rpc.NewClient(addr)

	assert.Nil(t, srv.Up())
	assert.Nil(t, cln.Connect())
	assert.Nil(t, cln.Connect())
	assert.Nil(t, cln.Disconnect())
	assert.Nil(t, cln.Disconnect())
	assert.Nil(t, srv.Down())
}

func TestClientCall(t *testing.T) {
	srv := rpc.NewServer(addr)
	cln := rpc.NewClient(addr)

	assert.Nil(t, srv.Register(&RPC{}))
	assert.Nil(t, srv.Up())
	assert.Nil(t, cln.Connect())
	assert.Nil(t, cln.Call("RPC.Call", struct{}{}, &struct{}{}))
	assert.Nil(t, cln.Disconnect())
	assert.Nil(t, srv.Down())
}

func TestClientReconnect(t *testing.T) {
	srv := rpc.NewServer(addr)
	cln := rpc.NewClient(addr)

	assert.Nil(t, srv.Register(&RPC{}))
	assert.Nil(t, srv.Up())
	assert.Nil(t, cln.Connect())
	assert.Nil(t, cln.Call("RPC.Call", struct{}{}, &struct{}{}))
	assert.Nil(t, cln.Disconnect())
	assert.NotNil(t, cln.Call("RPC.Call", struct{}{}, &struct{}{}))
	assert.Nil(t, cln.Connect())
	assert.Nil(t, cln.Call("RPC.Call", struct{}{}, &struct{}{}))
	assert.Nil(t, cln.Disconnect())
	assert.Nil(t, srv.Down())
}
