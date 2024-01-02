package net

import (
	"testing"

	"github.com/ayaskovets/consensus/pkg/net"
)

func TestClientGracefulShutdown(t *testing.T) {
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
