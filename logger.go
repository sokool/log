package log

import (
	"io"
	"log"
	"os"
)

var Default = New(os.Stdout, All)

type Option int64

const (
	Date Option = 1 << iota
	Time
	Type
	Tag
	Location
	Colors

	All = Date | Time | Type | Tag | Location | Colors
)

type Logger struct {
	writer   io.Writer
	verbose  bool
	tag      string
	option   Option
	location int
	handlers []Handler
}

func New(w io.Writer, o Option) *Logger {
	return &Logger{
		writer:  w,
		verbose: true,
		option:  o,
	}
}

func (l *Logger) Tag(name string) *Logger {
	n := l.new()
	n.tag = name
	return n
}

func (l *Logger) Handlers(h ...Handler) *Logger {
	n := l.new()
	n.handlers = h
	return n
}

func (l *Logger) Verbose(b bool) *Logger {
	n := l.new()
	n.verbose = b
	return n
}

func (l *Logger) Location(depth int) *Logger {
	n := l.new()
	n.location = depth
	return n
}

func (l *Logger) Options(o Option) *Logger {
	n := l.new()
	n.option = o
	return n
}

func (l *Logger) Write(p []byte) (n int, err error) {
	l.printf(string(p), 4)
	return len(p), nil
}

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
