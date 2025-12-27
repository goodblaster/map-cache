package resp

import (
	"fmt"
	"strconv"

	"github.com/tidwall/resp"
)

// RESP response constructors using tidwall/resp library

// SimpleString returns a RESP simple string (+OK\r\n)
func SimpleString(s string) resp.Value {
	return resp.SimpleStringValue(s)
}

// BulkString returns a RESP bulk string ($5\r\nhello\r\n)
func BulkString(s string) resp.Value {
	return resp.StringValue(s)
}

// Integer returns a RESP integer (:42\r\n)
func Integer(i int) resp.Value {
	return resp.IntegerValue(i)
}

// Array returns a RESP array (*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n)
func Array(values []resp.Value) resp.Value {
	return resp.ArrayValue(values)
}

// NullBulkString returns a RESP null bulk string ($-1\r\n)
func NullBulkString() resp.Value {
	return resp.NullValue()
}

// Error returns a RESP error (-ERR message\r\n)
func Error(msg string) resp.Value {
	return resp.ErrorValue(fmt.Errorf("%s", msg))
}

// OK returns +OK\r\n
func OK() resp.Value {
	return SimpleString("OK")
}

// Pong returns +PONG\r\n or bulk string with message
func Pong(message string) resp.Value {
	if message == "" {
		return SimpleString("PONG")
	}
	return BulkString(message)
}

// ConvertToRESP converts a Go value to RESP format
func ConvertToRESP(value any) resp.Value {
	if value == nil {
		return NullBulkString()
	}

	switch v := value.(type) {
	case string:
		return BulkString(v)
	case int:
		return Integer(v)
	case int64:
		return Integer(int(v))
	case float64:
		// Redis doesn't have native float type, return as string
		return BulkString(strconv.FormatFloat(v, 'f', -1, 64))
	case bool:
		if v {
			return Integer(1)
		}
		return Integer(0)
	case []any:
		// Convert array to RESP array
		values := make([]resp.Value, len(v))
		for i, item := range v {
			values[i] = ConvertToRESP(item)
		}
		return Array(values)
	case map[string]any:
		// Convert object to RESP array of key-value pairs
		values := make([]resp.Value, 0, len(v)*2)
		for key, val := range v {
			values = append(values, BulkString(key))
			values = append(values, ConvertToRESP(val))
		}
		return Array(values)
	default:
		// For unknown types, return as bulk string (JSON-encoded if needed)
		return BulkString(fmt.Sprintf("%v", v))
	}
}
