package logger

import (
	"log/slog"
	"os"
)

// NewLogger creates a new structured logger given the specified LoggerOutput, passing nil will silently ignore events
func NewLogger() *Logger {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

}

// Info prints an event to the info stream
func (log *Logger) Info(format string, v ...any) {
	log.infoLog.Printf(format+"\n", v...)
}

// Warn prints an event to the warn stream
func (log *Logger) Warn(format string, v ...any) {
	log.warnLog.Printf(format+"\n", v...)
}

// Error prints an event to the error stream
func (log *Logger) Error(format string, v ...any) {
	log.errLog.Printf(format+"\n", v...)
}

func (log *Logger) Fatal(format string, v ...any) {
	log.errLog.Fatalf(format+"\n", v...)
}

func (log *Logger) FatalErr(err error) {
	log.errLog.Fatalln(err)
}
