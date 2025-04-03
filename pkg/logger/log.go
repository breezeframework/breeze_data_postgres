package logger

import (
	"log/slog"
	"sync"
)

var (
	logInstance = slog.Default()
	mu          sync.Mutex
)

func SetLogger(l *slog.Logger) {
	mu.Lock()
	defer mu.Unlock()
	logInstance = l
}

func Logger() *slog.Logger {
	return logInstance
}
