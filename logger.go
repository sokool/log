package log

import (
	"io"
	"log"
	"os"
)

var Default = New(os.Stdout, All)

// Option ...
type Option int64

const (
	// Date render Message with date in 2006/01/02 format
	Date Option = 1 << iota

	// Time render Message with time in 15:04:05.000000 format
	Time

	// Type render Message with one of INF, DBG, ERR strings
	Type

	// Tag render Message with tag name
	Tag

	// Trace render Message with filename and line number
	Trace

	// Colors render Message with colors
	Colors

	// JSON makes output with json format instead text
	JSON

	All = Date | Time | Type | Tag | Trace | Colors
)

// Logger support three types(levels) of logging
//
// INF - default, when you want emphasize that something important (not negative)
// has happened. It should be used for informing about rare situation in your code,
// such as database connection has been established.
//
// DBG - it meant to be verbose, like every few lines of code, when you decide
// that part of your code implementation did something important for internal
// state of your library/code.
//
// ERR - when your code is a place where error is received but there is no
// good way of handling that situation you might log it
// todo
//   - default attributes data added to each Message
//   - log.Format option
type Logger struct {
	writer   io.Writer
	verbose  bool
	tag      string
	option   Option
	trace    int
	handlers []Handler
}

// New instance of logger
//
// io.Writer w will receive all messages that are generated internally from
// strings and arguments that are passed to Logger.Printf.
//
// Option o represents what is going to be included in strings that are passed
// into io.Writer
func New(w io.Writer, o ...Option) *Logger {
	if len(o) == 0 {
		o = append(o, All)
	}
	return &Logger{
		writer:  w,
		verbose: true,
		trace:   2,
		option:  o[0],
	}
}

// Tag creates new instance of Logger with predefined tag name
func (l *Logger) Tag(name string) *Logger {
	n := l.new()
	n.tag = name
	return n
}

// Handlers creates new instance of Logger and all Message's are passed into
// Handler just after it is writer to io.Writer
func (l *Logger) Handlers(h ...Handler) *Logger {
	n := l.new()
	n.handlers = h
	return n
}

// Verbose create new Logger instance, switch if DBG type should be also passed
// to io.Writer. It might be useful to keep it enabled in local, testing or
// staging environment but on production in some cases might be disabled
func (l *Logger) Verbose(enable bool) *Logger {
	n := l.new()
	n.verbose = enable
	return n
}

// Trace create new Logger instance
func (l *Logger) Trace(depth int) *Logger {
	n := l.new()
	n.trace = 2 + depth
	return n
}

// Options create new Logger instance
func (l *Logger) Options(o Option) *Logger {
	n := l.new()
	n.option = o
	return n
}

func (l *Logger) Infof(text string, args ...any) {
	l.write(text, "INF", args...)
}

func (l *Logger) Warnf(text string, args ...any) {
	l.write(text, "WRN", args...)
}

func (l *Logger) Debugf(text string, args ...any) {
	l.write(text, "DBG", args...)
}

func (l *Logger) Errorf(text string, args ...any) {
	l.write(text, "ERR", args...)
}

// Write tbd
func (l *Logger) Write(p []byte) (n int, err error) {
	l.write(string(p), "")
	return len(p), nil
}

// Printf
// when text is json format and has no arguments a then it will be transformed
func (l *Logger) Printf(text string, args ...any) {
	l.write(text, "", args...)
}

func (l *Logger) write(text, typ string, args ...any) {
	if l.tag != "" {
		text = l.tag + ":" + text
	}
	m := NewMessage(text, l.trace, args...)
	if typ != "" {
		m.typ = typ
	}
	if m.typ == "DBG" && !l.verbose {
		return
	}

	for _, rfn := range l.handlers {
		rfn(m)
	}
	if m.typ == "DBG" && !l.verbose {
		return
	}

	b, err := m.Render(l.option)
	if err != nil {
		log.Printf("sokool.log: message decode failed %s", err)
	}
	if _, err = l.writer.Write(append(b, '\n')); err != nil {
		log.Printf("sokool.log: message write failed %s", err)
	}
}

func (l *Logger) new() *Logger {
	return &Logger{
		writer:   l.writer,
		verbose:  l.verbose,
		tag:      l.tag,
		handlers: l.handlers,
		option:   l.option,
		trace:    l.trace,
	}
}

func Printf(format string, args ...any) {
	Default.write(format, "INF", args...)
}

func Debugf(format string, args ...any) {
	Default.write(format, "DBG", args...)
}

func Errorf(format string, args ...any) {
	Default.write(format, "ERR", args...)
}
