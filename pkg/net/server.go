package net

import (
	"log"
	"net"
	"net/rpc"
	"sync"
)

type Server struct {
	addr string
	rpc  *rpc.Server

	mu       sync.Mutex
	listener net.Listener
	wg       sync.WaitGroup
	closed   chan any
}

func NewServer(addr string) *Server {
	return &Server{
		addr:   addr,
		rpc:    rpc.NewServer(),
		closed: make(chan any),
	}
}

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

func (server *Server) Close() error {
	server.mu.Lock()
	defer server.mu.Unlock()

	close(server.closed)
	err := server.listener.Close()
	server.wg.Wait()
	return err
}
