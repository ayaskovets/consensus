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
	shutdown chan any

	wg sync.WaitGroup
}

// Constructs a new server object
func NewServer(addr string) *Server {
	return &Server{
		addr:     addr,
		rpc:      rpc.NewServer(),
		shutdown: make(chan any),
	}
}

// Returns the address that the server accepts incoming connection on
func (server *Server) Addr() string {
	return server.addr
}

// Registers the rcvr object as an RPC receiver
// Can be called multiple times
func (server *Server) Register(rcvr any) error {
	return server.rpc.Register(rcvr)
}

// Starts to accept incoming RPS requests
// Non-blocking
func (server *Server) Up() error {
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
				case <-server.shutdown:
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
func (server *Server) Down() error {
	server.mu.Lock()
	defer server.mu.Unlock()

	close(server.shutdown)
	err := server.listener.Close()
	server.wg.Wait()
	return err
}
