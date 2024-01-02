package net

import "testing"

func TestClientGracefulShutdown(t *testing.T) {
	srv := NewServer(":10000")
	if err := srv.Up(); err != nil {
		t.Fatal(err)
	}

	cln := NewClient(":10000")
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
	srv := NewServer(":10000")
	if err := srv.Up(); err != nil {
		t.Fatal(err)
	}

	cln := NewClient(":10000")
	if err := cln.Connect(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Connect(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := cln.Disconnect(); err != nil {
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
