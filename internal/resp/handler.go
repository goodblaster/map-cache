package resp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/goodblaster/map-cache/internal/config"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/tidwall/resp"
)

// CommandHandler is a function that handles a RESP command
type CommandHandler func(s *Session, args []resp.Value) error

// Command registry maps command names to their handlers
var commandRegistry = make(map[string]CommandHandler)

// RegisterCommand registers a command handler
func RegisterCommand(name string, handler CommandHandler) {
	commandRegistry[strings.ToUpper(name)] = handler
}

// HandleCommand dispatches a RESP command to its handler
func HandleCommand(s *Session, cmd resp.Value) error {
	// Commands must be arrays
	if cmd.Type() != resp.Array {
		return fmt.Errorf("expected array, got %s", cmd.Type())
	}

	values := cmd.Array()
	if len(values) == 0 {
		return fmt.Errorf("empty command")
	}

	// First element is the command name
	cmdName := strings.ToUpper(values[0].String())
	args := values[1:]

	// If in MULTI mode, queue commands (except EXEC, DISCARD, MULTI)
	if s.multiMode {
		if cmdName != "EXEC" && cmdName != "DISCARD" && cmdName != "MULTI" {
			s.multiCmds = append(s.multiCmds, cmd)
			return s.WriteValue(SimpleString("QUEUED"))
		}
	}

	// Look up the handler
	handler, exists := commandRegistry[cmdName]
	if !exists {
		return fmt.Errorf("unknown command '%s'", cmdName)
	}

	// Execute the handler
	return handler(s, args)
}

func init() {
	// Register generic commands
	RegisterCommand("PING", handlePing)
	RegisterCommand("ECHO", handleEcho)
	RegisterCommand("SELECT", handleSelect)
	RegisterCommand("COMMAND", handleCommand)
	RegisterCommand("HELLO", handleHello)
	RegisterCommand("CLIENT", handleClient)
	RegisterCommand("FLUSHDB", handleFlushDB)
	RegisterCommand("FLUSHALL", handleFlushAll)
}

// handlePing implements the PING command
func handlePing(s *Session, args []resp.Value) error {
	if len(args) == 0 {
		return s.WriteValue(Pong(""))
	}
	return s.WriteValue(Pong(args[0].String()))
}

// handleEcho implements the ECHO command
func handleEcho(s *Session, args []resp.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'echo' command")
	}
	return s.WriteValue(BulkString(args[0].String()))
}

// handleSelect implements the SELECT command (maps to cache selection)
func handleSelect(s *Session, args []resp.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'select' command")
	}

	// Parse database index
	dbIndex := args[0].String()

	// Map database index to cache name
	// db 0 -> "default" (only special case)
	// db 1 -> "1", db 2 -> "2", etc. (use numbers as cache names)
	var cacheName string
	if dbIndex == "0" {
		cacheName = config.RESPDefaultCache
	} else {
		// Use the number as the cache name directly
		// This means caches must be created with numeric names to be accessible via Redis
		cacheName = dbIndex
	}

	// Note: We don't validate if the cache exists here - it will be created on first use
	// or return an error when a command tries to use it
	s.selectedCache = cacheName

	return s.WriteOK()
}

// handleCommand implements the COMMAND command (returns info about commands)
// For now, we return a minimal response to satisfy client compatibility checks
func handleCommand(s *Session, args []resp.Value) error {
	if len(args) == 0 {
		// COMMAND with no args - return empty array for now
		// A full implementation would return info about all commands
		return s.WriteValue(Array([]resp.Value{}))
	}

	// COMMAND INFO, COMMAND COUNT, etc. - return minimal responses
	subCmd := strings.ToUpper(args[0].String())
	switch subCmd {
	case "COUNT":
		// Return number of registered commands
		return s.WriteValue(Integer(len(commandRegistry)))
	case "INFO":
		// Return info about specific commands - return null for now
		return s.WriteValue(NullBulkString())
	default:
		return s.WriteError(fmt.Sprintf("ERR unknown COMMAND subcommand '%s'", subCmd))
	}
}

// handleHello implements the HELLO command (RESP3 protocol negotiation)
// For now, we only support RESP2, so we return a minimal response
func handleHello(s *Session, args []resp.Value) error {
	// HELLO [protover [AUTH username password] [SETNAME clientname]]
	// We accept it but always use RESP2
	// Return server info as a map-like array
	response := []resp.Value{
		BulkString("server"), BulkString("map-cache"),
		BulkString("version"), BulkString("1.0.0"),
		BulkString("proto"), Integer(2), // We support RESP2
		BulkString("mode"), BulkString("standalone"),
		BulkString("role"), BulkString("master"),
	}
	return s.WriteValue(Array(response))
}

// handleClient implements the CLIENT command (client connection management)
// For now, we support only minimal subcommands
func handleClient(s *Session, args []resp.Value) error {
	if len(args) == 0 {
		return s.WriteError("ERR wrong number of arguments for 'client' command")
	}

	subCmd := strings.ToUpper(args[0].String())
	switch subCmd {
	case "SETINFO":
		// CLIENT SETINFO - silently accept but ignore
		return s.WriteOK()
	case "GETNAME":
		// CLIENT GETNAME - return null for now
		return s.WriteValue(NullBulkString())
	case "SETNAME":
		// CLIENT SETNAME - silently accept but ignore
		return s.WriteOK()
	case "ID":
		// CLIENT ID - return connection ID
		return s.WriteValue(Integer(int(s.connID)))
	default:
		return s.WriteError(fmt.Sprintf("ERR unknown CLIENT subcommand '%s'", subCmd))
	}
}

// handleFlushDB implements the FLUSHDB command (delete all keys in current database/cache)
func handleFlushDB(s *Session, args []resp.Value) error {
	// FLUSHDB [ASYNC|SYNC] - we ignore ASYNC/SYNC for now (always synchronous)
	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("FLUSHDB")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 30*time.Second)
	defer cancel()

	// Get all keys using wildcard pattern
	keys := cache.WildKeys(ctx, "*")

	// Delete each key
	for _, key := range keys {
		cache.Delete(ctx, key)
	}

	return s.WriteOK()
}

// handleFlushAll implements the FLUSHALL command (delete all keys in all databases/caches)
// For map-cache, we interpret this as flushing only the currently selected cache
// (same as FLUSHDB) since we don't have a way to iterate all caches
func handleFlushAll(s *Session, args []resp.Value) error {
	// FLUSHALL [ASYNC|SYNC] - we ignore ASYNC/SYNC for now (always synchronous)
	// Note: In Redis, this flushes ALL databases. In map-cache, we only flush
	// the current cache since there's no global cache iterator available.
	return handleFlushDB(s, args)
}
