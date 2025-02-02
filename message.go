package log

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
)

type Message struct {
	tags       []string
	level      Level
	text       string
	file       string
	funk       string
	line       int
	args       []any
	createdAt  time.Time
	attributes []int
}

func NewMessage(text string, deep int, args ...any) Message {
	m := Message{text: text, args: args, createdAt: time.Now(), level: INFO}
	if len(m.args) == 1 {
		if _, ok := m.args[0].(error); ok {
			m.level = ERROR
		}
	}

	if i := m.index(m.text, false); i > 0 {
		for _, s := range strings.Split(text[:i], ":") {
			l, ok := levels[strings.TrimSpace(s)]
			if s == "" {
				continue
			}
			if ok {
				m.level = l
				continue
			}
			m.tags = append(m.tags, s)
		}

		if m.text = strings.Replace(m.text, text[:i], "", 1); m.text[0] == ':' {
			m.text = m.text[1:]
		}
	}

	m.text = strings.TrimSpace(m.text)
	if i := m.index(m.text, true); i >= 0 {
		var s Data
		if json.Unmarshal([]byte(m.text[i:]), &s) == nil {
			m.args, m.text = append(m.args, s), m.text[:i]+"%v"
		}
	}

	if n := len(m.args); n > 0 {
		for i, s := range strings.Split(m.text, "%") {
			if n < i {
				break
			}
			if len(s) < 1 {
				continue
			}
			if s[0:1] == "v" {
				m.attributes = append(m.attributes, i-1)
			}
		}
	}

	m.text = strings.ReplaceAll(m.text, "%#v", "%s")
	// todo call it when Trace option is enabled
	if p, n, l, ok := runtime.Caller(deep + 1); ok {
		m.file, m.line, m.funk = n, l, runtime.FuncForPC(p).Name()
	}

	return m
}

func (m Message) Render(o Option) ([]byte, error) {
	// todo decide based on Option what fields should be attached to json output
	if o&JSON != 0 {
		return m.MarshalJSON()
	}

	var s string
	var c = o&Colors != 0
	if o&Date != 0 {
		s += fmt.Sprintf("%s ", m.createdAt.Format("2006/01/02"))
	}
	if o&Time != 0 {
		s += fmt.Sprintf("%s ", m.createdAt.Format("15:04:05.000000"))
	}
	if o&Levels != 0 {
		s += fmt.Sprintf("[%s] ", m.level.Render(true, c))
	}
	if t := m.Tag(c); o&Tags != 0 && t != "" {
		s += fmt.Sprintf("[%s] ", t)
	}
	s += m.Text(c, o&Properties != 0)

	if o&Trace != 0 {
		if l := m.Location(c); l != "" {
			s += fmt.Sprintf(" %s", l)
		}
	}

	return []byte(strings.TrimSpace(s)), nil
}

func (m Message) Text(colors, properties bool) string {
	var n = len(m.args)
	var args []any
	for i := range m.args {
		args = append(args, m.args[i])
	}
	for _, i := range m.attributes {
		if i > n {
			break
		}
		if !properties {
			args[i] = ""
			continue
		}
		switch f := args[i].(type) {
		case Data:
			args[i] = f.properties(colors)
		case map[string]any:
			args[i] = Data(f).properties(colors)
		default:
			var s Data
			b, _ := json.Marshal(f)
			json.Unmarshal(b, &s)
			args[i] = s.properties(colors)
		}
	}
	return strings.ReplaceAll(fmt.Sprintf(m.text, args...), "  ", " ")
}

func (m Message) Location(colors bool) string {
	i := strings.LastIndex(m.file, "/")
	s := fmt.Sprintf("%s:%d", m.file[i+1:], m.line)
	if colors {
		return fmt.Sprintf("\x1b[35;1m%s\x1b[0m", s)
	}
	return s
}

func (m Message) Tag(colors bool) string {
	s := strings.Join(m.tags, ":")
	if colors && s != "" {
		return fmt.Sprintf("\x1b[34;1m%s\x1b[0m", s)
	}
	return s
}

func (m Message) Type(colors bool) string {
	return m.level.Render(true, colors)
}

func (m Message) CreatedAt() time.Time {
	return m.createdAt
}

func (m Message) String() string {
	b, _ := m.Render(Date | Time | Tags | Levels)
	return string(b)
}

func (m Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Fields())
}

func (m Message) MarshalText() ([]byte, error) {
	return []byte(m.Fields().properties(false)), nil
}

func (m Message) Fields() Data {
	var a []any
	var n = len(m.args)
	var t string
	for _, i := range m.attributes {
		if i > n {
			break
		}
		a = append(a, m.args[i])
	}

	for i := range m.tags {
		t += fmt.Sprintf("%s", strings.Title(m.tags[i]))
	}
	return Data{
		"tag":   t,
		"tags":  m.tags,
		"level": m.level.String(),
		"text":  m.Text(false, false),
		"file":  m.file,
		"func":  m.funk,
		"line":  m.line,
		"date":  m.createdAt,
		"attr":  a,
	}
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

type Data map[string]any

func (a Data) properties(color bool) string {
	var s string
	for n, v := range a {
		switch x := v.(type) {
		case string:
			if strings.Contains(x, " ") {
				v = fmt.Sprintf(`"%s"`, v)
			}
		case fmt.Stringer:
			v = fmt.Sprintf(`"%s"`, v)
		}

		if color {
			n = fmt.Sprintf("\u001B[90;1m%s\u001B[0m", n)
			v = fmt.Sprintf("\u001B[37;3m%v\u001B[0m", v)
		}
		s += fmt.Sprintf("%s=%v ", n, v)
	}
	return strings.TrimSpace(s)
}

//func (a Data) String() string {
//	b, _ := json.MarshalIndent(a, "", "\t")
//	return string(b)
//}

type color struct {
	r, g, b int
}

func newColor(r, g, b int) color {
	return color{r, g, b}
}

func (c color) text(s string) string {
	return fmt.Sprintf("\x1b[38;5;%d;%d;%d1m%s", c.r, c.g, c.b, s)
}
