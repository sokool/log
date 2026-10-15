package log

import (
	"encoding/json"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"time"
)

type Message struct {
	Tags       []string
	Level      Level
	TraceID    string
	text       string
	File       string
	Func       string
	Line       int
	ARGS       []any
	CreatedAt  time.Time
	properties Data
	format     format
}

func NewMessage(text string, deep int, args ...any) Message {
	m := Message{text: text, ARGS: args, CreatedAt: time.Now(), Level: INFO}
	if len(m.ARGS) == 1 {
		if _, ok := m.ARGS[0].(error); ok {
			m.Level = ERROR
		}
	}

	if i := m.index(m.text); i > 0 {
		for _, s := range strings.Split(text[:i], ":") {
			l, ok := levels[strings.TrimSpace(s)]
			if s == "" {
				continue
			}
			if ok {
				m.Level = l
				continue
			}
			m.Tags = append(m.Tags, s)
		}

		if m.text = strings.Replace(m.text, text[:i], "", 1); m.text[0] == ':' {
			m.text = m.text[1:]
		}
	}

	m.text = strings.TrimSpace(m.text)
	m.properties = Data{}
	m.format = format{text: m.text, args: append([]any(nil), args...)}
	m.format.prepare(m.properties)
	m.extractTraceID()
	// todo call it when Trace option is enabled
	if p, n, l, ok := runtime.Caller(deep + 1); ok {
		m.File, m.Line, m.Func = n, l, runtime.FuncForPC(p).Name()
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
		s += fmt.Sprintf("%s ", m.CreatedAt.Format("2006/01/02"))
	}
	if o&Time != 0 {
		s += fmt.Sprintf("%s ", m.CreatedAt.Format("15:04:05.000000"))
	}
	if o&Levels != 0 {
		s += fmt.Sprintf("[%s] ", m.Level.Render(true, c))
	}
	if t := m.Tag(c); o&Tags != 0 && t != "" {
		s += fmt.Sprintf("[%s] ", t)
	}
	s += m.Text(c, o&Properties != 0)

	if o&Trace != 0 {
		s = strings.TrimRight(s, " ")
		if id := m.traceID(c); id != "" {
			s += " " + id
		}
		if l := m.Location(c); l != "" {
			s += fmt.Sprintf(" %s", l)
		}
	}

	return []byte(strings.TrimSpace(s)), nil
}

func (m Message) Text(colors, properties bool) string {
	text := m.format.render(colors, properties)
	if !properties && len(m.format.fields) > 0 {
		return strings.TrimSpace(strings.ReplaceAll(text, "  ", " "))
	}
	return text
}

func (m Message) Location(colors bool) string {
	i := strings.LastIndex(m.File, "/")
	s := fmt.Sprintf("%s:%d", m.File[i+1:], m.Line)
	if colors {
		return fmt.Sprintf("\x1b[35;1m%s\x1b[0m", s)
	}
	return s
}

func (m Message) Tag(colors bool) string {
	s := strings.Join(m.Tags, ":")
	if colors && s != "" {
		return fmt.Sprintf("\x1b[34;1m%s\x1b[0m", s)
	}
	return s
}

func (m Message) Type(colors bool) string {
	return m.Level.Render(true, colors)
}

func (m Message) Properties() Data {
	var t string

	for i := range m.Tags {
		t += fmt.Sprintf("%s", strings.Title(m.Tags[i]))
	}
	data := Data{
		"tag":   t,
		"tags":  m.Tags,
		"level": m.Level.String(),
		"text":  m.Text(false, false),
		"file":  m.File,
		"func":  m.Func,
		"line":  m.Line,
		"date":  m.CreatedAt,
	}
	if m.TraceID != "" {
		data["trace_id"] = m.TraceID
	}
	var conflicts Data
	for k, v := range m.properties {
		_, reserved := data[k]
		if reserved || k == "attr" || k == "trace_id" {
			if conflicts == nil {
				conflicts = Data{}
			}
			conflicts[k] = v
			continue
		}
		data[k] = v
	}
	if len(conflicts) > 0 {
		data["attr"] = conflicts
	}
	return data
}

func (m *Message) extractTraceID() {
	names := contexts.traceNames()
	if len(names) == 0 {
		for name := range m.properties {
			normalized := strings.Map(func(r rune) rune {
				if r == '_' || r == '-' || r == '.' {
					return -1
				}
				return r
			}, strings.ToLower(name))
			if strings.HasSuffix(normalized, "traceid") && name != "trace_id" {
				names = append(names, name)
			}
		}
		slices.Sort(names)
		names = append([]string{"trace_id"}, names...)
	}
	for _, name := range names {
		id, ok := m.properties[name].(string)
		if !ok || id == "" {
			continue
		}
		m.TraceID = id
		delete(m.properties, name)
		for _, a := range m.format.fields {
			delete(a.data, name)
		}
		return
	}
}

func (m Message) traceID(colors bool) string {
	if m.TraceID == "" {
		return ""
	}
	id := []rune(m.TraceID)
	id = id[:min(8, len(id))]
	text := "#" + string(id)
	if colors {
		return "\x1b[36;1m" + text + "\x1b[0m"
	}
	return text
}

func (m Message) index(text string) int {
	if strings.HasPrefix(text, `⇨`) || strings.HasPrefix(text, "\n") {
		return 0
	}
	if i := strings.IndexAny(text, " \t{["); i >= 0 {
		text = text[:i]
	}
	return strings.LastIndex(text, ":")
}

func (m Message) String() string {
	b, _ := m.Render(Date | Time | Tags | Levels | Properties)
	return string(b)
}

func (m Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Properties())
}

func (m Message) MarshalText() ([]byte, error) {
	return []byte(m.Properties().properties(false)), nil
}
