package test

import (
	"testing"

	"github.com/ayaskovets/consensus/pkg/net"
)

func TestGracefulShutdown(t *testing.T) {
	srv := net.NewServer(":10000")
	if err := srv.Up(); err != nil {
		t.Fatal(err)
	}

	cln := net.NewClient(":10000")
	if err := cln.Connect(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Disconnect(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := srv.Down(); err != nil {
		t.Fatal(err)
	}
}

func TestClientIdempotency(t *testing.T) {
	srv := net.NewServer(":10000")
	if err := srv.Up(); err != nil {
		t.Fatal(err)
	}

	cln := net.NewClient(":10000")
	for i := 0; i < 2; i++ {
		if err := cln.Connect(); err != nil {
			t.Log(err)
			t.Fail()
		}
	}

	for i := 0; i < 2; i++ {
		if err := cln.Disconnect(); err != nil {
			t.Log(err)
			t.Fail()
		}
	}

	if err := srv.Down(); err != nil {
		t.Fatal(err)
	}
}

func TestClientCall(t *testing.T) {
	srv := net.NewServer(":10000")
	if err := srv.Register(&PingRPC{}); err != nil {
		t.Fatal(err)
	}

	if err := srv.Up(); err != nil {
		t.Fatal(err)
	}

	cln := net.NewClient(":10000")
	if err := cln.Connect(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Call("PingRPC.Ping", PingArgs{}, &PingReply{}); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Disconnect(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := srv.Down(); err != nil {
		t.Fatal(err)
	}
}

func TestClientReconnect(t *testing.T) {
	srv := net.NewServer(":10000")
	if err := srv.Register(&PingRPC{}); err != nil {
		t.Fatal(err)
	}

	if err := srv.Up(); err != nil {
		t.Fatal(err)
	}

	cln := net.NewClient(":10000")
	if err := cln.Connect(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Call("PingRPC.Ping", PingArgs{}, &PingReply{}); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Disconnect(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Call("PingRPC.Ping", PingArgs{}, &PingReply{}); err == nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Connect(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Call("PingRPC.Ping", PingArgs{}, &PingReply{}); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Disconnect(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := srv.Down(); err != nil {
		t.Fatal(err)
	}
}
