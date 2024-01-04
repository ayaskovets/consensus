package rpc

import (
	"log"
	"net"
	"net/rpc"
	"sync"
)

// Wrapper for RPC server
type Server struct {
	addr net.Addr
	rpc  *rpc.Server

	mu       sync.Mutex
	listener net.Listener
	shutdown chan any

	wg sync.WaitGroup
}

// Construct new server object
func NewServer(addr net.Addr) *Server {
	server := Server{
		addr: addr,
		rpc:  rpc.NewServer(),

		mu:       sync.Mutex{},
		listener: nil,
		shutdown: make(chan any),

		wg: sync.WaitGroup{},
	}
	close(server.shutdown)
	return &server
}

// Register rcvr object as RPC receiver.
// Having multiple receivers of different types is allowed
func (server *Server) Register(rcvr any) error {
	return server.rpc.Register(rcvr)
}

// Begin to accept incoming RPS requests.
// Non-blocking
//
// Idempotent. Returns nil if already running i.e. has not been stopped since
// the last startup. Each call to this function must be followed by a
// corresponding shutdown
func (server *Server) Up() error {
	server.mu.Lock()
	defer server.mu.Unlock()

	select {
	case <-server.shutdown:
		server.shutdown = make(chan any)
	default:
		return nil
	}

	var err error
	server.listener, err = net.Listen(server.addr.Network(), server.addr.String())
	if err != nil {
		return err
	}
	log.Printf("listening on %s", server.addr)

	server.wg.Add(1)
	go func() {
		defer server.wg.Done()
		for {
			conn, err := server.listener.Accept()
			if err != nil {
				select {
				case <-server.shutdown:
					return
				default:
					log.Fatalf("error while accepting on %s: %s", server.addr, err)
					return
				}
			}

			server.wg.Add(1)
			go func() {
				defer server.wg.Done()
				server.rpc.ServeConn(conn)
			}()
		}
	}()

	return nil
}

// Shutdown RPC server. Blocks until all connections to the server are
// terminated. Successful call indicates that server can be started again
//
// Idempotent. Returns nil if already stopped
func (server *Server) Down() error {
	server.mu.Lock()
	defer server.mu.Unlock()

	select {
	case <-server.shutdown:
		return nil
	default:
	}
	log.Printf("stop listening on %s", server.addr)

	close(server.shutdown)
	err := server.listener.Close()
	server.wg.Wait()
	return err
}
