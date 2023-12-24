package net

import (
	"log"
	"net"
	"net/rpc"
	"sync"
)

type Server struct {
	address string

	mu  sync.Mutex
	rpc *rpc.Server
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
		rpc:     nil,
	}
}

func (server *Server) Serve(rcvr any) {
	server.mu.Lock()

	server.rpc = rpc.NewServer()
	err := server.rpc.Register(&rcvr)
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", server.address)
	if err != nil {
		log.Fatal(err)
	}

	server.mu.Unlock()

	log.Printf("listening on %s", server.address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			server.rpc.ServeConn(conn)
		}()
	}
}
