package logging

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
)

const (
	callDepth = 3
	flags     = log.LstdFlags | log.LUTC | log.Lmicroseconds
)

type Logger struct {
	info *log.Logger
	err  *log.Logger
}

func (l *Logger) header() string {
	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		return "%s"
	}

	return fmt.Sprint(file, ":", strconv.Itoa(line), " %s")
}

func (l *Logger) Info(msg string) {
	l.info.Printf(l.header(), msg)
}

func (l *Logger) Error(err error) {
	if w := errors.Unwrap(err); w != nil {
		err = w
	}
	l.err.Printf(l.header(), err)
}

func (l *Logger) Fatal(err error) {
	l.Error(err)
	os.Exit(1)
}

func New(service string) *Logger {
	prefix := "[" + service + "] "
	return &Logger{
		info: log.New(os.Stdout, prefix, flags),
		err:  log.New(os.Stderr, prefix, flags),
	}
}
