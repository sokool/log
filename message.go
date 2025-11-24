package log

import (
	"encoding/json"
	"fmt"
	"iter"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	Tags       []string
	Level      Level
	text       string
	File       string
	Func       string
	Line       int
	ARGS       []any
	CreatedAt  time.Time
	attributes []int
}

func NewMessage(text string, deep int, args ...any) Message {
	m := Message{text: text, ARGS: args, CreatedAt: time.Now(), Level: INFO}
	if len(m.ARGS) == 1 {
		if _, ok := m.ARGS[0].(error); ok {
			m.Level = ERROR
		}
	}

	if i := m.index(m.text, false); i > 0 {
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
	if i := m.index(m.text, true); i >= 0 {
		var s Data
		if json.Unmarshal([]byte(m.text[i:]), &s) == nil {
			m.ARGS, m.text = append(m.ARGS, s), m.text[:i]+"%v"
		}
	}

	if n := len(m.ARGS); n > 0 {
		var c int
		for i := 0; i < len(m.text)-1; i++ {
			if m.text[i] != '%' {
				continue
			}
			if m.text[i+1] == 'v' {
				m.attributes = append(m.attributes, c)
			}
			c++
		}

	}

	m.text = strings.ReplaceAll(m.text, "%#v", "%s")
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
		if l := m.Location(c); l != "" {
			s += fmt.Sprintf(" %s", l)
		}
	}

	return []byte(strings.TrimSpace(s)), nil
}

func (m Message) Text(colors, properties bool) string {
	var n = len(m.ARGS)
	var args []any
	for i := range m.ARGS {
		args = append(args, m.ARGS[i])
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
		case string, int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64,
			complex64, complex128:
			args[i] = f
		case []byte:
			args[i] = string(f)
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

func (m Message) Properties() Data {
	var a []any
	var n = len(m.ARGS)
	var t string
	for _, i := range m.attributes {
		if i > n {
			break
		}
		a = append(a, m.ARGS[i])
	}

	for i := range m.Tags {
		t += fmt.Sprintf("%s", strings.Title(m.Tags[i]))
	}
	return Data{
		"tag":   t,
		"tags":  m.Tags,
		"level": m.Level.String(),
		"text":  m.Text(false, false),
		"file":  m.File,
		"func":  m.Func,
		"line":  m.Line,
		"date":  m.CreatedAt,
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

// properties returns a string of key=value pairs, optionally colored.
func (d Data) properties(color bool, delim ...string) string {
	var s strings.Builder

	for n, v := range d.Flat(delim...) {
		// Quote values if needed
		if strings.Contains(v, " ") {
			v = fmt.Sprintf(`"%s"`, v)
		}

		// Apply color if requested
		if color {
			n = fmt.Sprintf("\u001B[90;1m%s\u001B[0m", n)
			v = fmt.Sprintf("\u001B[37;3m%s\u001B[0m", v)
		}

		s.WriteString(fmt.Sprintf("%s=%v ", n, v))
	}

	return strings.TrimSpace(s.String())
}

// Flat returns an iterator of flattened key-value pairs, sorted by key.
func (d Data) Flat(delim ...string) iter.Seq2[string, string] {
	delimChar := "."
	if len(delim) > 0 && delim[0] != "" {
		delimChar = delim[0]
	}

	return func(yield func(string, string) bool) {
		d.walk("", d, delimChar, yield)
	}
}

// walk recursively flattens value v, sorting map keys at each level.
func (d Data) walk(p string, v any, delim string, yield func(string, string) bool) bool {
	switch x := v.(type) {

	case map[string]any:
		// 1. Collect keys
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		// 2. Sort keys for deterministic iteration
		slices.Sort(keys)

		// 3. Iterate over sorted keys
		for _, k := range keys {
			n := d.join(p, k, delim)
			if !d.walk(n, x[k], delim, yield) {
				return false
			}
		}

	case Data:
		// Same logic for Data (which is a distinct type from map[string]any in switches)
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		for _, k := range keys {
			n := d.join(p, k, delim)
			if !d.walk(n, x[k], delim, yield) {
				return false
			}
		}

	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			for i := 0; i < rv.Len(); i++ {
				elem := rv.Index(i).Interface()
				n := d.join(p, strconv.Itoa(i), delim)
				if !d.walk(n, elem, delim, yield) {
					return false
				}
			}
			return true
		}

		return yield(p, fmt.Sprint(v))
	}

	return true
}

func (d Data) join(a, b, delim string) string {
	if a == "" {
		return b
	}
	return a + delim + b
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

func isNumber(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		complex64, complex128:
		return true
	}
	return false
}
