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
// todo
//   - default attributes data added to each Message
//   - log.Format option
type Logger struct {
	writer   io.Writer
	verbose  bool
	tag      string
	option   Option
	location int
	human    bool
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
		human:   true,
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

func (l *Logger) Format(human bool) *Logger {
	l.human = human
	return l
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

func (l *Logger) Infof(text string, args ...any) {
	l.write(text, "INF", 4, args...)
}

func (l *Logger) Debugf(text string, args ...any) {
	l.write(text, "DBG", 4, args...)
}

func (l *Logger) Errorf(text string, args ...any) {
	l.write(text, "ERR", 4, args...)
}

// Write tbd
func (l *Logger) Write(p []byte) (n int, err error) {
	l.write(string(p), "", 4)
	return len(p), nil
}

// Printf
// when text is json format and has no arguments a then it will be transformed
func (l *Logger) Printf(text string, args ...any) {
	l.write(text, "", 4, args...)
}

func (l *Logger) write(text, typ string, depth int, args ...any) {
	if l.tag != "" {
		text = l.tag + ":" + text
	}
	m := NewMessage(text, args...)
	if typ != "" {
		m.typ = typ
	}
	if m.typ == "DBG" && !l.verbose {
		return
	}

	var b []byte
	var err error
	for _, rfn := range l.handlers {
		rfn(m)
	}
	if m.typ == "DBG" && !l.verbose {
		return
	}
	switch l.human {
	case false:
		b, err = m.MarshalJSON()
	case true:
		b = []byte(m.Render(l.option, depth+l.location))
	}
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
		location: l.location,
		human:    l.human,
	}
}

func Printf(format string, args ...any) {
	Default.write(format, "INF", 4, args...)
}

func Debugf(format string, args ...any) {
	Default.write(format, "DBG", 4, args...)
}

func Errorf(format string, args ...any) {
	Default.write(format, "ERR", 4, args...)
}
