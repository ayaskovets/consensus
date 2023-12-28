package test

import (
	"testing"

	"github.com/ayaskovets/consensus/pkg/net"
)

func TestGracefulShutdown(t *testing.T) {
	srv := net.NewServer(":10000")
	if err := srv.Serve(); err != nil {
		t.Fatal(err)
	}

	cln := net.NewClient(":10000")
	if err := cln.Dial(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Close(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := srv.Close(); err != nil {
		t.Fatal(err)
	}
}
