package resp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/goodblaster/map-cache/internal/config"
	"github.com/goodblaster/map-cache/internal/log"
)

// Server represents the RESP (Redis Protocol) TCP server
type Server struct {
	listener       net.Listener
	activeConns    int32
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	shutdownChan   chan struct{}
}

// NewServer creates a new RESP server
func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		ctx:          ctx,
		cancel:       cancel,
		shutdownChan: make(chan struct{}),
	}
}

// Start starts the RESP server
func (srv *Server) Start() error {
	listener, err := net.Listen("tcp", config.RESPAddress)
	if err != nil {
		return fmt.Errorf("failed to start RESP server: %w", err)
	}

	srv.listener = listener

	log.With("address", config.RESPAddress).Info("RESP server started")

	// Accept connections in a goroutine
	go srv.acceptConnections()

	return nil
}

// acceptConnections accepts incoming client connections
func (srv *Server) acceptConnections() {
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			select {
			case <-srv.ctx.Done():
				// Server is shutting down
				return
			default:
				log.WithError(err).Error("Error accepting RESP connection")
				continue
			}
		}

		// Check connection limit
		activeConns := atomic.LoadInt32(&srv.activeConns)
		if activeConns >= int32(config.RESPMaxConnections) {
			log.Warn("RESP connection limit reached, rejecting connection")
			conn.Close()
			continue
		}

		// Increment active connection count
		atomic.AddInt32(&srv.activeConns, 1)

		// Handle connection in a new goroutine
		srv.wg.Add(1)
		go srv.handleConnection(conn)
	}
}

// handleConnection handles a client connection
func (srv *Server) handleConnection(conn net.Conn) {
	defer srv.wg.Done()
	defer atomic.AddInt32(&srv.activeConns, -1)

	session := NewSession(conn)
	session.Handle()
}

// Shutdown gracefully shuts down the RESP server
func (srv *Server) Shutdown(ctx context.Context) error {
	log.Info("Shutting down RESP server...")

	// Stop accepting new connections
	srv.cancel()

	if srv.listener != nil {
		srv.listener.Close()
	}

	// Wait for all connections to finish (or context timeout)
	done := make(chan struct{})
	go func() {
		srv.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("RESP server shut down gracefully")
		return nil
	case <-ctx.Done():
		log.Warn("RESP server shutdown timeout, forcing close")
		return ctx.Err()
	}
}

// ActiveConnections returns the number of active connections
func (srv *Server) ActiveConnections() int {
	return int(atomic.LoadInt32(&srv.activeConns))
}
