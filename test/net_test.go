package test

import (
	"testing"
	"time"

	"github.com/ayaskovets/consensus/pkg/net"
)

func TestGracefulShutdown(t *testing.T) {
	srv := net.NewServer(":10000")
	go srv.Serve()
	time.Sleep(time.Millisecond * 100)

	cln := net.NewClient(":10000")
	err := cln.Dial()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	err = cln.Close()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	err = srv.Close()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}
