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
	var err error
	server.listener, err = net.Listen("tcp", server.addr)
	if err != nil {
		return err
	}
	log.Printf("listening on %s", server.addr)

	server.wg.Add(1)
	defer server.wg.Done()

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			select {
			case <-server.closed:
				log.Printf("stop listening on %s", server.addr)
				return nil
			default:
				return err
			}
		}

		go func() {
			server.wg.Add(1)
			defer server.wg.Done()

			server.rpc.ServeConn(conn)
		}()
	}
}

func (server *Server) Close() error {
	close(server.closed)
	err := server.listener.Close()
	server.wg.Wait()
	return err
}
