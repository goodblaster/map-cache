package log

import "github.com/goodblaster/logos"

// logosAdapter adapts the logos library to implement our Logger interface.
// This is the only file in the codebase (besides main.go) that imports logos.
type logosAdapter struct {
	// Wrap logos.Logger to capture contextual loggers
	// returned by logos.WithError() and logos.With()
	impl logos.Logger
}

// LogosAdapter creates a new Logger backed by a logos.Logger instance.
// This function should only be called from main.go during application startup.
func LogosAdapter(logosLogger logos.Logger) Logger {
	return &logosAdapter{impl: logosLogger}
}

func (l *logosAdapter) Info(a ...any) {
	l.impl.Info(a...)
}

func (l *logosAdapter) Infof(format string, args ...any) {
	l.impl.Infof(format, args...)
}

func (l *logosAdapter) Warn(a ...any) {
	l.impl.Warn(a...)
}

func (l *logosAdapter) Warnf(format string, args ...any) {
	l.impl.Warnf(format, args...)
}

func (l *logosAdapter) Error(a ...any) {
	l.impl.Error(a...)
}

func (l *logosAdapter) Errorf(format string, args ...any) {
	l.impl.Errorf(format, args...)
}

func (l *logosAdapter) Fatal(a ...any) {
	l.impl.Fatal(a...)
}

func (l *logosAdapter) WithError(err error) Logger {
	// logos.WithError returns a contextual logger
	return &logosAdapter{impl: l.impl.WithError(err)}
}

func (l *logosAdapter) With(key string, value any) Logger {
	// logos.With returns a contextual logger
	return &logosAdapter{impl: l.impl.With(key, value)}
}

func (l *logosAdapter) Print(a ...any) {
	l.impl.Print(a...)
}
