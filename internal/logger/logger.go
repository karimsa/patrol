package logger

import (
	"log"
	"os"
)

type LogLevel int

const (
	LevelNone  LogLevel = -1
	LevelInfo  LogLevel = 0
	LevelDebug LogLevel = 1
)

type Logger struct {
	level       LogLevel
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
}

func New(level LogLevel, prefix string) Logger {
	flags := log.LstdFlags | log.Lmsgprefix
	return Logger{
		level:       level,
		warnLogger:  log.New(os.Stderr, prefix+"warn: ", flags),
		debugLogger: log.New(os.Stderr, prefix+"debug: ", flags),
		infoLogger:  log.New(os.Stdout, prefix+"info: ", flags),
	}
}

func (l Logger) Warnf(msg string, vals ...interface{}) {
	l.warnLogger.Printf(msg, vals...)
}

func (l Logger) Infof(msg string, vals ...interface{}) {
	if l.level >= LevelInfo {
		l.infoLogger.Printf(msg, vals...)
	}
}

func (l Logger) Debugf(msg string, vals ...interface{}) {
	if l.level >= LevelDebug {
		l.debugLogger.Printf(msg, vals...)
	}
}
