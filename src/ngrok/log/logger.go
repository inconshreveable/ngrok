package log

import (
	log "code.google.com/p/log4go"
	"fmt"
)

const (
	logfile string = "ngrok.log"
)

func init() {
	// log4go automatically sets the global logger to write to stdout
	// and we don't want that by default
	delete(log.Global, "stdout")
}

func LogToConsole() {
	log.Global.AddFilter("log", log.DEBUG, log.NewConsoleLogWriter())
}

func LogToFile() {
	log.Global.AddFilter("log", log.DEBUG, log.NewFileLogWriter(logfile, true))
}

type Logger interface {
	AddLogPrefix(string)
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{}) error
	Error(string, ...interface{}) error
}

type PrefixLogger struct {
	*log.Logger
	prefix string
}

func NewPrefixLogger() Logger {
	return &PrefixLogger{Logger: &log.Global}
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
