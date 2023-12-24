package test

import (
	"testing"
)

func TestNodesConnectivity(t *testing.T) {
	cluster := NewCluster([]string{":10001", ":10002", ":10003"})
	err := cluster.Up()
	if err != nil {
		t.Fatal(err)
	}

	err = cluster.Down()
	if err != nil {
		t.Fatal(err)
	}
}
