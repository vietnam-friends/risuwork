package logger

import (
	"log/slog"
	"sync/atomic"
)

var playerLogger atomic.Pointer[slog.Logger]
var adminLogger atomic.Pointer[slog.Logger]

func init() {
	playerLogger.Store(slog.Default())
	adminLogger.Store(slog.Default())
}

func Player() *slog.Logger {
	return playerLogger.Load()
}

func Admin() *slog.Logger {
	return adminLogger.Load()
}

func SetPlayer(l *slog.Logger) {
	playerLogger.Store(l)
}

func SetAdmin(l *slog.Logger) {
	adminLogger.Store(l)
}
