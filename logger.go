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
	// Location render [Message] with filename with line number
	Location
	//Colors render Message with colors
	Colors
	All  = Date | Time | Type | Tag | Location | Colors
	None = 0
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
type Logger struct {
	writer   io.Writer
	verbose  bool
	tag      string
	option   Option
	location int
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

// Location create new Logger instance
func (l *Logger) Location(depth int) *Logger {
	n := l.new()
	n.location = depth
	return n
}

// Options create new Logger instance
func (l *Logger) Options(o Option) *Logger {
	n := l.new()
	n.option = o
	return n
}

// Write tbd
func (l *Logger) Write(p []byte) (n int, err error) {
	l.printf(string(p), 4)
	return len(p), nil
}

// Printf tbd
func (l *Logger) Printf(format string, a ...interface{}) {
	l.printf(format, 4, a...)
}

func (l *Logger) printf(format string, depth int, a ...interface{}) {
	if l.tag != "" {
		format = l.tag + ":" + format
	}
	m := NewMessage(format, a...)
	if m.typ == "DBG" && !l.verbose {
		return
	}
	s := m.Render(l.option, depth+l.location)
	if _, err := l.writer.Write([]byte(s + "\n")); err != nil {
		log.Printf("sokool:log write failed %s", err)
	}
	for _, rfn := range l.handlers {
		rfn(m)
	}
}

func (l *Logger) new() *Logger {
	return &Logger{
		writer:   l.writer,
		verbose:  l.verbose,
		tag:      l.tag,
		handlers: l.handlers,
		option:   l.option,
		location: l.location,
	}
}

func Printf(format string, args ...interface{}) {
	Default.printf(format, 4, args...)
}

func Debugf(format string, args ...interface{}) {
	Default.printf("dbg "+format, 4, args...)
}

func Errorf(format string, args ...interface{}) {
	Default.printf("err "+format, 4, args...)
}
