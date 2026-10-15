package log

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type format struct {
	text   string
	args   []any
	fields []*argument
	parts  []part
}

// prepare snapshots ordinary formatting and retains only property operands for rendering.
func (f *format) prepare(properties Data) {
	next, reordered := 0, false
	for i := 0; i < len(f.text); {
		start := i
		for i < len(f.text) && f.text[i] != '%' {
			i++
		}
		f.parts = append(f.parts, part{text: f.text[start:i]})
		if i == len(f.text) {
			break
		}
		i++
		start = i
		for i < len(f.text) && strings.ContainsRune("#0+- ", rune(f.text[i])) {
			i++
		}
		flags := f.text[start:i]
		var spec strings.Builder
		spec.WriteString("%" + flags)
		indexed := false
		var operands []any
		for dimension := 0; dimension < 2; dimension++ {
			if f.index(&i, &next) {
				reordered, indexed = true, true
			}
			start = i
			if i < len(f.text) && f.text[i] == '*' {
				if next < len(f.args) {
					operands = append(operands, f.args[next])
				}
				next++
				i++
				indexed = false
			} else {
				for i < len(f.text) && f.text[i] >= '0' && f.text[i] <= '9' {
					i++
				}
			}
			spec.WriteString(f.text[start:i])
			if dimension == 1 || i == len(f.text) || f.text[i] != '.' {
				break
			}
			spec.WriteByte('.')
			i++
		}
		if !indexed && f.index(&i, &next) {
			reordered, indexed = true, true
		}
		verb, size := utf8.DecodeRuneInString(f.text[i:])
		if i == len(f.text) {
			verb, size = 0, 0
		}
		i += size
		p := part{}
		if verb != '%' && verb != 0 {
			if next < len(f.args) {
				value := f.args[next]
				if verb == 'v' && !strings.Contains(flags, "#") {
					a := &argument{value: value, explicit: strings.Contains(flags, "+")}
					a.prepare()
					if a.data != nil || a.explicit {
						for k, v := range a.data {
							properties[k] = v
						}
						p.argument, value = a, a
						f.fields = append(f.fields, a)
					}
				}
				operands = append(operands, value)
			} else if indexed {
				fmt.Fprintf(&spec, "[%d]", len(operands)+1)
			}
			next++
		}
		if verb != 0 {
			spec.WriteRune(verb)
		}
		p.text, p.args = spec.String(), operands
		if p.argument == nil {
			p.text, p.args = fmt.Sprintf(p.text, p.args...), nil
		}
		f.parts = append(f.parts, p)
	}
	if !reordered && next < len(f.args) {
		empty := ""
		f.parts = append(f.parts, part{text: fmt.Sprintf(empty, f.args[next:]...)})
	}
	f.text, f.args = "", nil
}

func (f *format) index(i, next *int) bool {
	if *i >= len(f.text) || f.text[*i] != '[' {
		return false
	}
	end := strings.IndexByte(f.text[*i:], ']')
	if end < 0 {
		return false
	}
	n, err := strconv.Atoi(f.text[*i+1 : *i+end])
	if err != nil || n < 1 {
		return false
	}
	*next = n - 1
	*i += end + 1
	return true
}

func (f *format) render(colors, show bool) string {
	var out strings.Builder
	for _, p := range f.parts {
		out.WriteString(p.render(colors, show))
	}
	return out.String()
}
