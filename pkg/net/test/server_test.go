package net

import (
	"testing"

	"github.com/ayaskovets/consensus/pkg/net"
)

func TestServerGracefulShutdown(t *testing.T) {
	srv := net.NewServer(":10000")
	if err := srv.Up(); err != nil {
		t.Fatal(err)
	}

	if err := srv.Down(); err != nil {
		t.Fatal(err)
	}
}

func TestServerIdempotency(t *testing.T) {
	srv := net.NewServer(":10000")
	if err := srv.Up(); err != nil {
		t.Fatal(err)
	}

	if err := srv.Up(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := srv.Down(); err != nil {
		t.Fatal(err)
	}

	if err := srv.Down(); err != nil {
		t.Fatal(err)
	}
}

func TestServerRestart(t *testing.T) {
	srv := net.NewServer(":10000")
	if err := srv.Up(); err != nil {
		t.Fatal(err)
	}

	if err := srv.Down(); err != nil {
		t.Fatal(err)
	}

	if err := srv.Up(); err != nil {
		t.Fatal(err)
	}

	if err := srv.Down(); err != nil {
		t.Fatal(err)
	}
}
