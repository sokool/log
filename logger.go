package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

var Default = New(os.Stdout, All)

type Logger struct {
	writer   io.Writer
	verbose  bool
	colors   bool
	tag      string
	option   Option
	handlers []Handler
}

func New(w io.Writer, o Option) *Logger {
	return &Logger{
		writer:  w,
		colors:  w == os.Stdout,
		verbose: true,
		option:  o,
	}
}

func (l *Logger) Tag(name string) *Logger {
	return &Logger{
		writer:   l.writer,
		verbose:  l.verbose,
		colors:   l.colors,
		tag:      name,
		handlers: l.handlers,
		option:   l.option,
	}
}

func (l *Logger) Handlers(h ...Handler) *Logger {
	return &Logger{
		writer:   l.writer,
		verbose:  l.verbose,
		colors:   l.colors,
		tag:      l.tag,
		option:   l.option,
		handlers: h,
	}
}

func (l *Logger) Verbose(b bool) *Logger {
	return &Logger{
		writer:   l.writer,
		verbose:  b,
		colors:   l.colors,
		tag:      l.tag,
		handlers: l.handlers,
		option:   l.option,
	}
}

func (l *Logger) Write(p []byte) (n int, err error) {
	l.Printf(string(p))
	return len(p), nil
}

func (l *Logger) Printf(format string, a ...interface{}) {
	m := NewMessage(format, a...)
	if l.tag != "" && m.tag != "" {
		m.tag = fmt.Sprintf("%s:%s", l.tag, m.tag)
	}

	if _, err := l.writer.Write([]byte(m.Render(l.option) + "\n")); err != nil {
		log.Printf("sokool:log write failed %s", err)
	}
	for _, rfn := range l.handlers {
		rfn(m)
	}
}

func Printf(format string, args ...interface{}) {
	Default.Printf(format, args...)
}
