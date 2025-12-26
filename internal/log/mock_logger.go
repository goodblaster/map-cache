package log

import "sync"

// MockLogger is a test logger that captures log messages for assertions.
// Use this in tests instead of importing logos directly.
type MockLogger struct {
	mu       sync.Mutex
	Messages []string
	Errors   []error
	Fields   map[string]any
}

// NewMockLogger creates a new mock logger for testing.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		Messages: []string{},
		Errors:   []error{},
		Fields:   make(map[string]any),
	}
}

func (m *MockLogger) Info(a ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, "INFO")
}

func (m *MockLogger) Infof(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, "INFO: "+format)
}

func (m *MockLogger) Warn(a ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, "WARN")
}

func (m *MockLogger) Warnf(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, "WARN: "+format)
}

func (m *MockLogger) Error(a ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, "ERROR")
}

func (m *MockLogger) Errorf(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, "ERROR: "+format)
}

func (m *MockLogger) Fatal(a ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, "FATAL")
	// Don't actually exit in tests
}

func (m *MockLogger) WithError(err error) Logger {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors = append(m.Errors, err)
	return m
}

func (m *MockLogger) With(key string, value any) Logger {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Fields[key] = value
	return m
}

func (m *MockLogger) Print(a ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, "PRINT")
}

// HasMessage checks if a message was logged (useful for test assertions).
func (m *MockLogger) HasMessage(substr string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, msg := range m.Messages {
		if contains(msg, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
