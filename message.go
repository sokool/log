package log

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
)

type Message struct {
	tag        []string
	typ        string
	text       string
	args       []any
	createdAt  time.Time
	attributes Struct
}

var types = map[string]bool{"dbg": true, "err": true, "inf": true}

func NewMessage(text string, args ...any) Message {
	m := Message{text: text, args: args, createdAt: time.Now(), typ: "INF"}
	if len(m.args) == 1 {
		if _, ok := m.args[0].(error); ok {
			m.typ = "ERR"
		}
	}

	if i := m.index(m.text, false); i > 0 {
		for _, s := range strings.Split(text[:i], ":") {
			switch s = strings.TrimSpace(s); {
			case types[s]:
				m.typ = strings.ToTitle(s)
			case s != "":
				m.tag = append(m.tag, s)
			}
		}

		if m.text = strings.Replace(m.text, text[:i], "", 1); m.text[0] == ':' {
			m.text = m.text[1:]
		}
	}

	m.text = strings.TrimSpace(m.text)
	if i := m.index(m.text, true); i >= 0 {
		var s map[string]any
		if json.Unmarshal([]byte(m.text[i:]), &s) == nil {
			m.args, m.text = append(m.args, s), m.text[:i]+"%v"
		}
	}

	for i := range m.args {
		if f, ok := m.args[i].(map[string]any); ok {
			m.attributes = f
		}
	}

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
	if t := m.Tag(c); o&Tag != 0 && t != "" {
		s += fmt.Sprintf("[%s] ", t)
	}

	s += m.Text(c)

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

func (m Message) Text(colors bool) string {
	for i := range m.args {
		if f, ok := m.args[i].(map[string]any); ok {
			m.args[i] = Struct(f).render(colors)
		}
	}
	return fmt.Sprintf(m.text, m.args...)
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
	s := strings.Join(m.tag, ":")
	if colors && s != "" {
		return fmt.Sprintf("\x1b[36;1m%s\x1b[0m", s)
	}
	return s
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

func (m Message) String() string {
	return m.Render(Date | Time | Tag | Type)
}

func (m Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Fields())
}

func (m Message) MarshalText() ([]byte, error) {
	return []byte(m.Fields().render(false)), nil
}

func (m Message) Fields() Struct {
	var f = Struct{
		"tag":       m.Tag(false),
		"type":      m.typ,
		"text":      m.Text(false),
		"location":  m.Location(false, 5),
		"createdAt": m.createdAt,
	}
	for i := range m.attributes {
		f[i] = m.attributes[i]
	}
	return f
}

func (m Message) index(text string, js bool) int {
	var i int
	if len(text) == 0 {
		return 0
	}
	if strings.HasPrefix(text, `⇨`) || strings.HasPrefix(text, "\n") {
		return 0
	}
	if text[0] == '\n' {
		return 0
	}
	if i = strings.LastIndex(text, ":"); i > 0 {
		if n := strings.Index(text[:i], " "); n != -1 && n < i {
			i = n
		}
	}

	var b []byte
	var p int
	if p = strings.Index(text, "{"); p == -1 {
		p = strings.Index(text, "[")
	}
	if b = []byte(text); p == -1 || (!json.Valid(b[p:])) {
		return i
	}
	if js {
		return p
	}
	if i > p {
		return p
	}
	return i
}

type Struct map[string]any

func (f Struct) render(color bool) string {
	var s string
	for n, v := range f {
		switch x := v.(type) {
		case string:
			if strings.Contains(x, " ") {
				v = fmt.Sprintf(`"%s"`, v)
			}
		case fmt.Stringer:
			v = fmt.Sprintf(`"%s"`, v)
		}

		if color {
			n = fmt.Sprintf("\u001B[4m\x1b[35;1m%s\x1b[0m\u001B[24m", n)
		}
		s += fmt.Sprintf("%s=%v ", n, v)
	}
	return s
}

type color struct {
	r, g, b int
}

func newColor(r, g, b int) color {
	return color{r, g, b}
}

func (c color) text(s string) string {
	return fmt.Sprintf("\x1b[38;5;%d;%d;%d1m%s", c.r, c.g, c.b, s)
}
