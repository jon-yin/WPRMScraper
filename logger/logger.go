package logger

import (
	"io"
	"log"
	"os"
)

var (
	// DefaultLoggerOpts specifies a logger which logs warn events to stdout and error events to stderr
	DefaultLoggerOpts = &LoggerOutput{
		InfoOut: io.Discard,
		WarnOut: os.Stdout,
		ErrOut:  os.Stderr,
	}
	// Verbose LoggerOpts
	VerboseLoggerOpts = &LoggerOutput{
		InfoOut: os.Stdout,
		WarnOut: os.Stdout,
		ErrOut:  os.Stderr,
	}
)

// LoggerOutput specifies where different log levels should log out to
type LoggerOutput struct {
	InfoOut io.Writer // Where to write info events to
	WarnOut io.Writer // Where to write warn events to
	ErrOut  io.Writer // Where to write error events to
}

// Logger logs events of 3 levels; info, warn, and error
type Logger struct {
	infoLog *log.Logger
	warnLog *log.Logger
	errLog  *log.Logger
}

// NewLogger creates a new Logger given the specified LoggerOutput, passing nil will silently ignore events
func NewLogger(logOpts *LoggerOutput) *Logger {
	if logOpts == nil {
		logOpts = &LoggerOutput{
			InfoOut: io.Discard,
			WarnOut: io.Discard,
			ErrOut:  io.Discard,
		}
	}
	opts := &Logger{
		infoLog: log.New(logOpts.InfoOut, "INFO:", log.LstdFlags|log.Lshortfile),
		warnLog: log.New(logOpts.WarnOut, "WARN:", log.LstdFlags|log.Lshortfile),
		errLog:  log.New(logOpts.ErrOut, "ERROR:", log.LstdFlags|log.Lshortfile),
	}
	return opts
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

func (log *Logger) Fatalf(format string, v ...any) {
	log.errLog.Fatalf(format+"\n", v...)
}

func (log *Logger) Fatal(msg any) {
	log.errLog.Fatalln(msg)
}
