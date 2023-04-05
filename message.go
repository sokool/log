package log

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
)

type Message struct {
	tag       string
	typ       string
	text      string
	createdAt time.Time
}

var types = []string{"dbg", "err", "inf"}

func NewMessage(format string, args ...any) Message {
	m := Message{
		text:      fmt.Sprintf(format, args...),
		createdAt: time.Now(),
	}
	for _, n := range types {
		if p := strings.Index(m.text, n); p != -1 && (p == 0 || m.text[p-1] == ':') {
			m.text, m.typ = strings.Replace(m.text, n, "", 1), strings.ToTitle(n)
			break
		}
	}
	if len(args) != 0 && m.typ == "" {
		if _, ok := args[0].(error); ok {
			m.typ = "ERR"
		}
	}
	if m.typ == "" {
		m.typ = "INF"
	}
	if p := strings.LastIndex(m.text, ":"); p != -1 && !strings.Contains(m.text[0:p], " ") {
		m.text, m.tag = m.text[p+1:], m.text[0:p]
	}
	m.text = strings.TrimSpace(m.text)

	return m
}

func (m Message) Render(o Option, depth ...int) string {
	var s string
	var c = o&Colors != 0
	if o&Date != 0 {
		s += fmt.Sprintf("%s ", m.createdAt.Format("2006/01/02"))
	}
	if o&Time != 0 {
		s += fmt.Sprintf("%s ", m.createdAt.Format("15:04:05.000000"))
	}
	if o&Type != 0 {
		s += fmt.Sprintf("[%s] ", m.Type(c))
	}
	if o&Tag != 0 && m.tag != "" {
		s += fmt.Sprintf("[%s] ", m.Tag(c))
	}

	s += m.text

	if o&Location != 0 {
		if len(depth) == 0 {
			depth = append(depth, 2)
		}
		if l := m.Location(c, depth[0]); l != "" {
			s += fmt.Sprintf(" %s", l)
		}
	}
	return strings.TrimSpace(s)
}

func (m Message) String() string {
	return m.Render(Date | Time | Tag | Type)
}

func (m Message) Text() string {
	return m.text
}

func (m Message) Location(colors bool, depth int) string {
	if _, file, line, ok := runtime.Caller(depth); ok {
		i := strings.LastIndex(file, "/")
		s := fmt.Sprintf("%s:%d", file[i+1:], line)
		if colors {
			return fmt.Sprintf("\x1b[35;1m%s\x1b[0m", s)
		}
		return s
	}
	return ""
}

func (m Message) Tag(colors bool) string {
	if colors {
		return fmt.Sprintf("\x1b[36;1m%s\x1b[0m", m.tag)
	}
	return m.tag
}

func (m Message) Type(colors bool) string {
	if !colors {
		return m.typ
	}
	color := "%s"
	if colors {
		switch m.typ {
		case "INF":
			color = "\x1b[32;1m%s\x1b[0m" // green
		case "ERR":
			color = "\x1b[31;1m%s\x1b[0m" // red
		case "DBG":
			color = "\x1b[33;1m%s\x1b[0m" // yellow
		}
	}
	return fmt.Sprintf(color, m.typ)
}

func (m Message) CreatedAt() time.Time {
	return m.createdAt
}

func (m Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}
