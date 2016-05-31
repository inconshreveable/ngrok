package log

import (
	"fmt"
	log "github.com/alecthomas/log4go"
)

var root log.Logger = make(log.Logger)

func LogTo(target string, level_name string) {
	var writer log.LogWriter = nil

	switch target {
	case "stdout":
		writer = log.NewConsoleLogWriter()
	case "none":
		// no logging
	default:
		writer = log.NewFileLogWriter(target, true)
	}

	if writer != nil {
		var level = log.DEBUG

		switch level_name {
		case "FINEST":
			level = log.FINEST
		case "FINE":
			level = log.FINE
		case "DEBUG":
			level = log.DEBUG
		case "TRACE":
			level = log.TRACE
		case "INFO":
			level = log.INFO
		case "WARNING":
			level = log.WARNING
		case "ERROR":
			level = log.ERROR
		case "CRITICAL":
			level = log.CRITICAL
		default:
			level = log.DEBUG
		}

		root.AddFilter("log", level, writer)
	}
}

type Logger interface {
	AddLogPrefix(string)
	ClearLogPrefixes()
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{}) error
	Error(string, ...interface{}) error
}

type PrefixLogger struct {
	*log.Logger
	prefix string
}

func NewPrefixLogger(prefixes ...string) Logger {
	logger := &PrefixLogger{Logger: &root}

	for _, p := range prefixes {
		logger.AddLogPrefix(p)
	}

	return logger
}

func (pl *PrefixLogger) pfx(fmtstr string) interface{} {
	return fmt.Sprintf("%s %s", pl.prefix, fmtstr)
}

func (pl *PrefixLogger) Debug(arg0 string, args ...interface{}) {
	pl.Logger.Debug(pl.pfx(arg0), args...)
}

func (pl *PrefixLogger) Info(arg0 string, args ...interface{}) {
	pl.Logger.Info(pl.pfx(arg0), args...)
}

func (pl *PrefixLogger) Warn(arg0 string, args ...interface{}) error {
	return pl.Logger.Warn(pl.pfx(arg0), args...)
}

func (pl *PrefixLogger) Error(arg0 string, args ...interface{}) error {
	return pl.Logger.Error(pl.pfx(arg0), args...)
}

func (pl *PrefixLogger) AddLogPrefix(prefix string) {
	if len(pl.prefix) > 0 {
		pl.prefix += " "
	}

	pl.prefix += "[" + prefix + "]"
}

func (pl *PrefixLogger) ClearLogPrefixes() {
	pl.prefix = ""
}

// we should never really use these . . . always prefer logging through a prefix logger
func Debug(arg0 string, args ...interface{}) {
	root.Debug(arg0, args...)
}

func Info(arg0 string, args ...interface{}) {
	root.Info(arg0, args...)
}

func Warn(arg0 string, args ...interface{}) error {
	return root.Warn(arg0, args...)
}

func Error(arg0 string, args ...interface{}) error {
	return root.Error(arg0, args...)
}
