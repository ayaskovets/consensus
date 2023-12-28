package net

import (
	"log"
	"net"
	"net/rpc"
	"sync"
)

// Wrapper for an RPC server
type Server struct {
	addr string
	rpc  *rpc.Server

	mu       sync.Mutex
	listener net.Listener
	closed   chan any

	wg sync.WaitGroup
}

// Constructs a new server object
// To start to accept incoming RPS calls, call Serve()
func NewServer(addr string) *Server {
	return &Server{
		addr:   addr,
		rpc:    rpc.NewServer(),
		closed: make(chan any),
	}
}

// Starts to accept incoming RPS requests
// Non-blocking
func (server *Server) Serve() error {
	server.mu.Lock()
	defer server.mu.Unlock()

	var err error
	server.listener, err = net.Listen("tcp", server.addr)
	if err != nil {
		return err
	}
	log.Printf("listening on %s", server.addr)

	go func() {
		server.wg.Add(1)
		defer server.wg.Done()
		for {
			conn, err := server.listener.Accept()
			if err != nil {
				select {
				case <-server.closed:
					log.Printf("stop listening on %s", server.addr)
					return
				default:
					log.Fatalf("error while accepting on %s", server.addr)
					return
				}
			}

			go func() {
				server.wg.Add(1)
				defer server.wg.Done()

				server.rpc.ServeConn(conn)
			}()
		}
	}()

	return nil
}

// Shuts down the RPC server
func (server *Server) Close() error {
	server.mu.Lock()
	defer server.mu.Unlock()

	close(server.closed)
	err := server.listener.Close()
	server.wg.Wait()
	return err
}
