package node

import (
	"github.com/ayaskovets/consensus/pkg/net"
)

type Node struct {
	server *net.Server
	peers  map[string]*net.Client
}

func NewNode(addr string) *Node {
	return &Node{
		server: net.NewServer(addr),
		peers:  make(map[string]*net.Client),
	}
}

func (node *Node) Call(addr string, serviceMethod string, args any, reply any) error {
	return node.peers[addr].Call(serviceMethod, args, reply)
}

func (node *Node) Connect(addr string) error {
	peer := node.peers[addr]
	if peer != nil {
		return nil
	}

	client := net.NewClient(addr)
	err := client.Dial()
	if err != nil {
		return err
	}

	node.peers[addr] = client
	return nil
}

func (node *Node) Disconnect(addr string) error {
	peer := node.peers[addr]
	if peer == nil {
		return nil
	}

	err := peer.Close()
	if err != nil {
		return err
	}

	delete(node.peers, addr)
	return nil
}

func (node *Node) Up() error {
	return node.server.Serve()
}

func (node *Node) Down() error {
	for _, peer := range node.peers {
		err := peer.Close()
		if err != nil {
			return err
		}
	}
	return node.server.Close()
}
