package consensus

import (
	"testing"
	"time"

	"github.com/ayaskovets/consensus/pkg/net"
)

func TestGracefulShutdown(t *testing.T) {
	addr := ":10000"

	srv := net.NewServer(addr)
	go srv.Serve()
	time.Sleep(time.Millisecond * 100)

	cln := net.NewClient(addr)
	err := cln.Dial()
	if err != nil {
		t.Fail()
	}

	err = cln.Close()
	if err != nil {
		t.Fail()
	}

	err = srv.Close()
	if err != nil {
		t.Fail()
	}
}
