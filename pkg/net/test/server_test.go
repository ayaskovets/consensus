package test

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
	for i := 0; i < 2; i++ {
		if err := srv.Up(); err != nil {
			t.Log(err)
			t.Fail()
		}
	}

	for i := 0; i < 2; i++ {
		if err := srv.Down(); err != nil {
			t.Log(err)
			t.Fail()
		}
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

func TestServerRegister(t *testing.T) {
	srv := net.NewServer(":10000")
	if err := srv.Register(&PingRPC{}); err != nil {
		t.Fatal(err)
	}
}
