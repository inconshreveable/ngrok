package ngrok

import (
	log "code.google.com/p/log4go"
	"fmt"
)

var Log log.Logger

func init() {
	Log = make(log.Logger)
	//    Log.AddFilter("log", log.DEBUG, log.NewFileLogWriter("ngrok.log", true))
}

func LogToConsole() {
	Log.AddFilter("log", log.DEBUG, log.NewConsoleLogWriter())
}

func LogToFile() {
	Log.AddFilter("log", log.DEBUG, log.NewFileLogWriter("ngrok.log", true))
}

type Logger interface {
	AddLogPrefix(string)
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{}) error
	Error(string, ...interface{}) error
}

type PrefixLogger struct {
	prefix string
}

func NewPrefixLogger() Logger {
	return &PrefixLogger{}
}

func (pl *PrefixLogger) pfx(fmtstr string) interface{} {
	return fmt.Sprintf("%s %s", pl.prefix, fmtstr)
}

func (pl *PrefixLogger) Debug(arg0 string, args ...interface{}) {
	Log.Debug(pl.pfx(arg0), args...)
}

func (pl *PrefixLogger) Info(arg0 string, args ...interface{}) {
	Log.Info(pl.pfx(arg0), args...)
}

func (pl *PrefixLogger) Warn(arg0 string, args ...interface{}) error {
	return Log.Warn(pl.pfx(arg0), args...)
}

func (pl *PrefixLogger) Error(arg0 string, args ...interface{}) error {
	return Log.Error(pl.pfx(arg0), args...)
}

func (pl *PrefixLogger) AddLogPrefix(prefix string) {
	if len(pl.prefix) > 0 {
		pl.prefix += " "
	}

	pl.prefix += "[" + prefix + "]"
}
