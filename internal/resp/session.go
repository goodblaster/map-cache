package resp

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/goodblaster/map-cache/internal/config"
	"github.com/goodblaster/map-cache/internal/log"
	"github.com/tidwall/resp"
)

var connIDCounter uint64

// Session represents a client connection to the RESP server
type Session struct {
	conn          net.Conn
	connID        uint64
	reader        *resp.Conn
	selectedCache string
	multiMode     bool
	multiCmds     []resp.Value
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewSession creates a new session for a client connection
func NewSession(conn net.Conn) *Session {
	connID := atomic.AddUint64(&connIDCounter, 1)
	ctx, cancel := context.WithCancel(context.Background())

	return &Session{
		conn:          conn,
		connID:        connID,
		reader:        resp.NewConn(conn),
		selectedCache: config.RESPDefaultCache,
		multiMode:     false,
		multiCmds:     []resp.Value{},
		ctx:           ctx,
		cancel:        cancel,
	}
}

// ReadCommand reads a RESP command from the client
func (s *Session) ReadCommand() (resp.Value, error) {
	val, _, err := s.reader.ReadValue()
	return val, err
}

// WriteValue writes a RESP value to the client
func (s *Session) WriteValue(val resp.Value) error {
	return s.reader.WriteValue(val)
}

// WriteError writes a RESP error to the client
func (s *Session) WriteError(msg string) error {
	return s.WriteValue(Error(msg))
}

// WriteOK writes +OK to the client
func (s *Session) WriteOK() error {
	return s.WriteValue(OK())
}

// Close closes the session and connection
func (s *Session) Close() error {
	s.cancel()
	return s.conn.Close()
}

// Tag generates a unique tag for cache locking
func (s *Session) Tag(cmdName string) string {
	return fmt.Sprintf("resp-%d-%s", s.connID, cmdName)
}

// SelectedCache returns the currently selected cache name
func (s *Session) SelectedCache() string {
	return s.selectedCache
}

// Context returns the session context
func (s *Session) Context() context.Context {
	return s.ctx
}

// Handle processes commands for this session
func (s *Session) Handle() {
	defer s.Close()

	log.With("conn_id", s.connID).
		With("remote_addr", s.conn.RemoteAddr().String()).
		Info("RESP client connected")

	for {
		cmd, err := s.ReadCommand()
		if err != nil {
			// Client disconnected or error reading
			if err.Error() != "EOF" {
				log.WithError(err).
					With("conn_id", s.connID).
					Warn("Error reading command")
			}
			break
		}

		// Handle the command
		if err := HandleCommand(s, cmd); err != nil {
			log.WithError(err).
				With("conn_id", s.connID).
				With("command", cmd.String()).
				Error("Error handling command")

			// Try to send error to client
			s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	}

	log.With("conn_id", s.connID).Info("RESP client disconnected")
}
